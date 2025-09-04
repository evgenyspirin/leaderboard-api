package ports

import (
	"context"

	"leaderboard-api/internal/domain/event"
)

type Scorer interface {
	RunScorerPool(ctx context.Context, size int)
	ClosePool(ctx context.Context)
	GetInputChan() chan event.Event
	GetOutChan() chan event.Event
}
