package services

import (
	"context"
	"testing"

	"leaderboard-api/internal/domain/leader"

	"github.com/stretchr/testify/require"
)

type mockLBMemory struct {
	topN   func(int) leader.Leaders
	rankOf func(string) (leader.Leader, bool)
}

func (m *mockLBMemory) TopN(n int) leader.Leaders {
	return m.topN(n)
}

func (m *mockLBMemory) RankOf(id string) (leader.Leader, bool) {
	return m.rankOf(id)
}

func (m *mockLBMemory) RunLBWorker(_ context.Context)    {}
func (m *mockLBMemory) StopRankWorker(_ context.Context) {}
func (m *mockLBMemory) All() leader.Leaders              { return make(leader.Leaders, 0) }

func TestLeaderboardService_GetBboard(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		expected leader.Leaders
	}{
		{
			name:  "Returns Top 2",
			limit: 2,
			expected: leader.Leaders{
				{TalentID: "t-1", Score: 100},
				{TalentID: "t-2", Score: 90},
			},
		},
		{
			name:     "Returns Empty",
			limit:    0,
			expected: leader.Leaders{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLBMemory{
				topN: func(n int) leader.Leaders {
					return tt.expected
				},
			}

			svc := NewLeaderboardService(mock)
			got, err := svc.GetBboard(context.Background(), tt.limit)
			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestLeaderboardService_GetRankByID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		mockOut  leader.Leader
		mockOk   bool
		expected leader.Leader
	}{
		{
			name:     "Found",
			id:       "t-1",
			mockOut:  leader.Leader{TalentID: "t-1", Score: 95},
			mockOk:   true,
			expected: leader.Leader{TalentID: "t-1", Score: 95},
		},
		{
			name:     "Not found",
			id:       "t-999",
			mockOut:  leader.Leader{},
			mockOk:   false,
			expected: leader.Leader{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLBMemory{
				rankOf: func(id string) (leader.Leader, bool) {
					require.Equal(t, tt.id, id)
					return tt.mockOut, tt.mockOk
				},
			}

			svc := NewLeaderboardService(mock)
			got, err := svc.GetRankByID(context.Background(), tt.id)
			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}
