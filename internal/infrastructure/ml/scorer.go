package ml

import (
	"context"
	"leaderboard-api/internal/domain/event"
	"math/rand/v2"
	"time"

	"go.uber.org/zap"
)

// bufferSize
// Good to have +/- same size as writers.
// - Reducing the blocking of writers
// - Aligning producer/consumer speeds(like in streaming)
// "Rely on metrics, not guesses."
var bufferSize = 1000

type OutputChan = chan event.Event

type Scorer struct {
	in  chan event.Event
	out OutputChan
	log *zap.Logger
}

func New(ctx context.Context, log *zap.Logger) *Scorer {
	sc := &Scorer{
		in:  make(chan event.Event, bufferSize),
		out: make(OutputChan, bufferSize),
		log: log,
	}

	return sc
}

// RunScorerPool - We are using "WorkerPool" concurrency pattern
// to create a pool of parallel processes then "Fan-In" pattern
// to send results into one channel.
func (s *Scorer) RunScorerPool(ctx context.Context, size int) {
	s.log.Info("starting ml pool")
	for i := 0; i < size; i++ {
		go s.worker(ctx)
	}
}

func (s *Scorer) ClosePool(ctx context.Context) {
	close(s.in)
	close(s.out)
	s.log.Info("ml scorer pool gracefully stopped")
}

func (s *Scorer) worker(ctx context.Context) {
	for evnt := range s.in {
		evnt.Score = s.score(ctx, evnt.RawMetric, evnt.Skill)
		s.out <- evnt
	}
}

// todo: What is really happening in ML process ??
func (s *Scorer) score(ctx context.Context, metric float64, skill string) float64 {
	time.Sleep(time.Duration(rand.IntN(150-80+1)+80) * time.Millisecond)
	return rand.Float64()*(200.0-0.0) + 0.0
}

func (s *Scorer) GetInputChan() chan event.Event { return s.in }
func (s *Scorer) GetOutChan() OutputChan         { return s.out }
