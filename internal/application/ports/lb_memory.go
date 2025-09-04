package ports

import (
	"context"

	"leaderboard-api/internal/domain/leader"
)

type LBMemory interface {
	RunLBWorker(ctx context.Context)
	StopRankWorker(ctx context.Context)
	TopN(n int) leader.Leaders
	RankOf(talentID string) (l leader.Leader, ok bool)
	All() leader.Leaders
}
