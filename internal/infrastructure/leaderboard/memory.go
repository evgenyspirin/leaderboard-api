package leaderboard

import (
	"context"
	"sync"

	"github.com/google/btree"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"leaderboard-api/internal/domain/leader"
	"leaderboard-api/internal/infrastructure/ml"
)

// Minimal size of a node.
// Affects the balance between tree height and node thickness.
// 8, 16, 32,64 - depends on the load on the system.
// More readers 32–64
// More writers 16–32
const defaultDegree = 16

type LBMemory struct {
	mu           sync.RWMutex
	in           ml.OutputChan
	log          *zap.Logger
	bestByTalent map[string]float64
	// Even after 100 million insertions(burst of writes), search remains
	// almost just as fast because a B-tree is a “wide and shallow” structure with
	// excellent cache locality and strict balancing.
	// This is why google/btree consistently provides
	// fast Get and Ascend operations under heavy in-memory workloads.
	tree    *btree.BTreeG[key]
	metrics *prometheus.CounterVec
}

type key struct {
	Score    float64
	TalentID string
}

func New(ctx context.Context, log *zap.Logger, in ml.OutputChan, metrics *prometheus.CounterVec) *LBMemory {
	lbm := &LBMemory{
		log:          log,
		bestByTalent: make(map[string]float64),
		tree:         btree.NewG[key](defaultDegree, less),
		in:           in,
		metrics:      metrics,
	}

	// also:
	// lbm.wakeUp()
	// lbm.backupWorker()

	return lbm
}

func (lbm *LBMemory) RunLBWorker(ctx context.Context) {
	lbm.log.Info("starting leaderboard worker")

	defer func() {
		lbm.log.Info("leaderboard worker gracefully stopped")
	}()

	for evnt := range lbm.in {
		lbm.updateIfBetter(leader.Leader{
			Rank:     0,
			TalentID: evnt.TalentID,
			Score:    evnt.Score,
		})
		lbm.metrics.WithLabelValues("accepted").Inc()
	}
}

func (lbm *LBMemory) StopRankWorker(ctx context.Context) {
	close(lbm.in)
	lbm.log.Info("leaderboard rank worker gracefully stopped")
}

// updateIfBetter -  O(log N)
func (lbm *LBMemory) updateIfBetter(l leader.Leader) (updated bool) {
	lbm.mu.Lock()
	defer lbm.mu.Unlock()

	old, ok := lbm.bestByTalent[l.TalentID]
	if ok && l.Score <= old {
		return false
	}
	if ok {
		lbm.tree.Delete(key{Score: old, TalentID: l.TalentID})
	}

	lbm.tree.ReplaceOrInsert(key{Score: l.Score, TalentID: l.TalentID})
	lbm.bestByTalent[l.TalentID] = l.Score

	return true
}

// TopN - O(log N + n)
func (lbm *LBMemory) TopN(n int) leader.Leaders {
	// we can also return cash of TOP10, TOP50, TOP100
	// and update it by N time

	lbm.mu.RLock()
	defer lbm.mu.RUnlock()

	if n <= 0 {
		return nil
	}
	ls := make(leader.Leaders, 0, n)
	i := 0
	lbm.tree.Descend(func(k key) bool {
		i++
		ls = append(ls, &leader.Leader{Rank: i, TalentID: k.TalentID, Score: k.Score})
		return i < n
	})

	return ls
}

// RankOf - O(log N + rank)
func (lbm *LBMemory) RankOf(talentID string) (l leader.Leader, ok bool) {
	lbm.mu.RLock()
	defer lbm.mu.RUnlock()

	l.TalentID = talentID
	l.Score, ok = lbm.bestByTalent[talentID]
	if !ok {
		return l, false
	}

	target := key{Score: l.Score, TalentID: talentID}
	i := 0
	found := false
	lbm.tree.Descend(func(k key) bool {
		i++
		if !found && k == target {
			found = true
			return false
		}
		return true
	})
	if !found {
		return l, false
	}
	l.Rank = i

	return l, true
}

// All - O(n) For possible future backups
func (lbm *LBMemory) All() leader.Leaders {
	lbm.mu.RLock()
	defer lbm.mu.RUnlock()

	ls := make(leader.Leaders, 0, lbm.tree.Len())
	rank := 0
	lbm.tree.Ascend(func(k key) bool {
		rank++
		ls = append(ls, &leader.Leader{Rank: rank, TalentID: k.TalentID, Score: k.Score})
		return true
	})

	return ls
}

// less - comparator that determines the overall order of keys in the tree
func less(a, b key) bool {
	if a.Score != b.Score {
		return a.Score < b.Score
	}

	return a.TalentID < b.TalentID
}
