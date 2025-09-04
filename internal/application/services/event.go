package services

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"

	"leaderboard-api/internal/application/ports"
	"leaderboard-api/internal/domain/event"
)

type EventService struct {
	cache   ports.Cache
	scorer  ports.Scorer
	metrics *prometheus.CounterVec
}

func NewEventService(
	cache ports.Cache,
	scorer ports.Scorer,
	metrics *prometheus.CounterVec,

) ports.EventService {
	return &EventService{
		cache:   cache,
		scorer:  scorer,
		metrics: metrics,
	}
}

func (es *EventService) Create(ctx context.Context, e *event.Event) (bool, error) {
	duplicate := es.cache.IsSet(e.EventID)
	if !duplicate {
		es.cache.Set(e.EventID)
		es.scorer.GetInputChan() <- *e
	}

	es.metrics.WithLabelValues("duplicate").Inc()
	return duplicate, nil
}

func (es *EventService) Seed(ctx context.Context, cnt int) {
	for _, val := range generateRandomEvents(cnt) {
		es.scorer.GetInputChan() <- val
	}
}

func generateRandomEvents(n int) []event.Event {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	skills := []string{"dribble", "shoot", "pass", "defense", "rebound", "speed", "stamina", "vision"}

	now := time.Now().UTC()
	events := make([]event.Event, n)

	for i := 0; i < n; i++ {
		raw := round(r.Float64() * 100.0) // 0..100.0
		events[i] = event.Event{
			EventID:   uuid.New(),
			TalentID:  fmt.Sprintf("t-%03d", r.Intn(500)+1), // t-001..t-500
			RawMetric: raw,
			Skill:     skills[r.Intn(len(skills))],
			TS:        randomPastTime(r, now, 90*24*time.Hour), // last 90 days
			Score:     round(raw / 100.0),
		}
	}
	return events
}

func randomPastTime(r *rand.Rand, ref time.Time, window time.Duration) time.Time {
	delta := time.Duration(r.Int63n(int64(window)))
	t := ref.Add(-delta)
	return t.Truncate(time.Minute)
}

func round(v float64) float64 {
	return math.Round(v*100) / 100
}
