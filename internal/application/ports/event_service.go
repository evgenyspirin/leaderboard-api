package ports

import (
	"context"

	"leaderboard-api/internal/domain/event"
)

type EventService interface {
	Create(ctx context.Context, event *event.Event) (bool, error)
	Seed(ctx context.Context, cnt int)
}
