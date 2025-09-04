package services

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"

	"leaderboard-api/internal/domain/event"
)

type mockCache struct {
	isSetFunc func(id uuid.UUID) bool
	setCalls  map[uuid.UUID]int
}

func (m *mockCache) IsSet(id uuid.UUID) bool {
	if m.isSetFunc != nil {
		return m.isSetFunc(id)
	}
	return false
}

func (m *mockCache) Set(id uuid.UUID) {
	if m.setCalls == nil {
		m.setCalls = make(map[uuid.UUID]int)
	}
	m.setCalls[id]++
}

type mockScorer struct {
	ch chan event.Event
}

func (m *mockScorer) RunScorerPool(ctx context.Context, size int) {}
func (m *mockScorer) GetOutChan() chan event.Event                { return make(chan event.Event) }
func (m *mockScorer) GetInputChan() chan event.Event              { return m.ch }
func (m *mockScorer) ClosePool(ctx context.Context)               {}

func TestEventService_Create(t *testing.T) {
	t.Parallel()

	type tc struct {
		name            string
		isDuplicate     bool
		expectSetCalled bool
		expectSent      bool
	}
	cases := []tc{
		{
			name:            "new event, set in cache and send to scorer; duplicate=false",
			isDuplicate:     false,
			expectSetCalled: true,
			expectSent:      true,
		},
		{
			name:            "duplicate event, no set, no send; duplicate=true",
			isDuplicate:     true,
			expectSetCalled: false,
			expectSent:      false,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ev := &event.Event{
				EventID:  uuid.New(),
				TalentID: "t-001",
				TS:       time.Now().UTC(),
				Skill:    "shoot",
				Score:    0.42,
			}

			cache := &mockCache{
				isSetFunc: func(id uuid.UUID) bool { return tt.isDuplicate },
				setCalls:  make(map[uuid.UUID]int),
			}
			inCh := make(chan event.Event, 1)
			sc := &mockScorer{ch: inCh}

			metrics := prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "leaderboard",
					Subsystem: "ingest",
					Name:      "events_processed_total",
					Help:      "Total number of processed events",
				},
				[]string{"result"},
			)

			reg := prometheus.NewRegistry()
			require.NoError(t, reg.Register(metrics))

			svc := NewEventService(cache, sc, metrics)

			dup, err := svc.Create(context.Background(), ev)

			require.NoError(t, err)
			require.Equal(t, tt.isDuplicate, dup)

			if tt.expectSetCalled {
				require.Equal(t, 1, cache.setCalls[ev.EventID], "cache.Set trigger once")
			} else {
				require.Zero(t, cache.setCalls[ev.EventID], "cache.Set no trigger for duplicates")
			}

			if tt.expectSent {
				select {
				case got := <-inCh:
					require.Equal(t, *ev, got)
				default:
					t.Fatalf("An event was expected in the scorer input, but the channel is empty")
				}
			} else {
				select {
				case <-inCh:
					t.Fatalf("An unexpected event occurred in the scorer input for the duplicate")
				default:
				}
			}

			got := testutil.ToFloat64(metrics.WithLabelValues("duplicate"))
			require.Equal(t, float64(1), got, "The duplicate metric should be incremented once per Create call")
		})
	}
}

func TestRound(t *testing.T) {
	t.Parallel()

	type tc struct {
		in   float64
		want float64
	}
	cases := []tc{
		{in: 0, want: 0},
		{in: 1, want: 1},
		{in: 1.234, want: 1.23},
		{in: 1.235, want: 1.24},
		{in: 99.999, want: 100.00},
		{in: -1.234, want: -1.23},
		{in: -1.235, want: -1.24},
		{in: 12.3456, want: 12.35},
		{in: 12.3449, want: 12.34},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(funcName(tt.in), func(t *testing.T) {
			t.Parallel()
			got := round(tt.in)
			require.InDelta(t, tt.want, got, 1e-9)
		})
	}
}

func funcName(v float64) string { return "round_" + format(v) }
func format(v float64) string   { return strconv.FormatFloat(v, 'f', -1, 64) }
