package ports

import (
	"context"

	"leaderboard-api/internal/domain/leader"
)

type LeaderboardService interface {
	GetBboard(ctx context.Context, limit int) (leader.Leaders, error)
	GetRankByID(ctx context.Context, id string) (leader.Leader, error)
}
