package ml

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"leaderboard-api/internal/domain/event"
)

func TestScorer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		metric     float64
		skill      string
		workerSize int
	}{
		{"Basic test A", 42.0, "dribble", 1},
		{"Basic test B", 95.3, "jump", 3},
		{"Basic test C", 10.0, "shoot", 2},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			logger := zap.NewNop()

			sc := New(ctx, logger)

			sc.RunScorerPool(ctx, tt.workerSize)

			ev := event.Event{
				RawMetric: tt.metric,
				Skill:     tt.skill,
			}

			sc.GetInputChan() <- ev

			select {
			case out := <-sc.GetOutChan():
				require.InDelta(t, tt.metric, out.RawMetric, 0.001)
				require.Equal(t, tt.skill, out.Skill)
				require.GreaterOrEqual(t, out.Score, 0.0)
				require.LessOrEqual(t, out.Score, 200.0)
			case <-time.After(1 * time.Second):
				t.Fatal("Timeout waiting for scored event")
			}

			sc.ClosePool(ctx)
		})
	}
}
