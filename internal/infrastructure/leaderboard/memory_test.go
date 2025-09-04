package leaderboard

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"leaderboard-api/internal/domain/event"
	"leaderboard-api/internal/domain/leader"
)

func newTestLB(t *testing.T) *LBMemory {
	t.Helper()
	log := zaptest.NewLogger(t)
	return New(context.Background(), log, make(chan event.Event, 10), prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{}))
}

func TestUpdateIfBetter_Table(t *testing.T) {
	lb := newTestLB(t)

	tests := []struct {
		name     string
		talentID string
		score    float64
		want     bool
	}{
		{"First insert", "talent1", 50, true},
		{"Lower score ignored", "talent1", 40, false},
		{"Higher score updated", "talent1", 70, true},
		{"Another new insert", "talent2", 60, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok := lb.updateIfBetter(leader.Leader{
				TalentID: tt.talentID,
				Score:    tt.score,
			})
			require.Equal(t, tt.want, ok)
		})
	}
}

func TestTopN_Table(t *testing.T) {
	lb := newTestLB(t)

	_ = lb.updateIfBetter(leader.Leader{TalentID: "t1", Score: 10})
	_ = lb.updateIfBetter(leader.Leader{TalentID: "t2", Score: 30})
	_ = lb.updateIfBetter(leader.Leader{TalentID: "t3", Score: 20})

	tests := []struct {
		name       string
		limit      int
		wantLength int
		wantTop    string
	}{
		{"Top 1", 1, 1, "t2"},
		{"Top 2", 2, 2, "t2"},
		{"Top 3", 3, 3, "t2"},
		{"Top all", 5, 3, "t2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			top := lb.TopN(tt.limit)
			require.Len(t, top, tt.wantLength)
			if tt.wantLength > 0 {
				require.Equal(t, tt.wantTop, top[0].TalentID)
			}
		})
	}
}

func TestRankOf_Table(t *testing.T) {
	lb := newTestLB(t)

	_ = lb.updateIfBetter(leader.Leader{TalentID: "t1", Score: 15})
	_ = lb.updateIfBetter(leader.Leader{TalentID: "t2", Score: 25})
	_ = lb.updateIfBetter(leader.Leader{TalentID: "t3", Score: 35})

	tests := []struct {
		name      string
		talentID  string
		wantFound bool
		wantRank  int
	}{
		{"Existing - top", "t3", true, 1},
		{"Existing - second", "t2", true, 2},
		{"Non-existing", "unknown", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, ok := lb.RankOf(tt.talentID)
			require.Equal(t, tt.wantFound, ok)
			if ok {
				require.Equal(t, tt.wantRank, r.Rank)
			}
		})
	}
}

func TestAll_Table(t *testing.T) {
	lb := newTestLB(t)

	_ = lb.updateIfBetter(leader.Leader{TalentID: "t1", Score: 1})
	_ = lb.updateIfBetter(leader.Leader{TalentID: "t2", Score: 2})
	_ = lb.updateIfBetter(leader.Leader{TalentID: "t3", Score: 3})

	tasks := lb.All()

	require.Len(t, tasks, 3)

	expected := []string{"t1", "t2", "t3"}
	for i, task := range tasks {
		require.Equal(t, expected[i], task.TalentID)
	}
}
