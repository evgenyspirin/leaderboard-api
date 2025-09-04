package services

import (
	"context"

	"leaderboard-api/internal/application/ports"
	"leaderboard-api/internal/domain/leader"
)

type LeaderboardService struct {
	memory ports.LBMemory
}

func NewLeaderboardService(
	memory ports.LBMemory,
) ports.LeaderboardService {
	return &LeaderboardService{
		memory: memory,
	}
}

func (ls *LeaderboardService) GetBboard(ctx context.Context, limit int) (leader.Leaders, error) {
	return ls.memory.TopN(limit), nil
}

func (ls *LeaderboardService) GetRankByID(ctx context.Context, id string) (leader.Leader, error) {
	l, ok := ls.memory.RankOf(id)
	if !ok {
		return leader.Leader{}, nil
	}

	return l, nil
}
