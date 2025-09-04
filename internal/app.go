package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"leaderboard-api/config"
	"leaderboard-api/internal/application/services"
	"leaderboard-api/internal/infrastructure/cache"
	"leaderboard-api/internal/infrastructure/leaderboard"
	"leaderboard-api/internal/infrastructure/metrics"
	"leaderboard-api/internal/infrastructure/ml"
	"leaderboard-api/internal/interface/api/rest"
	"leaderboard-api/internal/interface/api/rest/middleware"
)

// poolSize - Always better to make up a decision based on measurement traffic,
// profiling, metrics and hardware.
// "Rely on metrics, not guesses."
var poolSize = 1000

type App struct {
	logger   *zap.Logger
	cfg      config.Config
	httpSrv  *http.Server
	mux      *http.ServeMux
	cache    *cache.Cache
	scorer   *ml.Scorer
	lbMemory *leaderboard.LBMemory
	metrics  *prometheus.CounterVec
}

func NewApp(ctx context.Context) (*App, error) {
	// logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("cannot initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// config
	if err := godotenv.Load(".env"); err != nil {
		logger.Fatal("error loading .env file", zap.Error(err))
	}
	cfg := config.Load()

	// router
	m := http.NewServeMux()
	// httpServer
	httpSrv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: middleware.RequestLog(logger)(m),
	}

	// cache
	c := cache.New(ctx, logger)
	// ml scorer
	s := ml.New(ctx, logger)
	// metrics
	mtr := metrics.New()
	// leaderboard memory
	lbMem := leaderboard.New(ctx, logger, s.GetOutChan(), mtr)

	return &App{
		logger:   logger,
		cfg:      cfg,
		httpSrv:  httpSrv,
		mux:      m,
		cache:    c,
		scorer:   s,
		lbMemory: lbMem,
		metrics:  mtr,
	}, nil
}

func (a *App) Close() {
	if a.logger != nil {
		_ = a.logger.Sync()
	}
}

// Run - The central place to launch and manage our application and
// parallel processes through a single context.
func (a *App) Run(ctx context.Context) error {
	// context with os signals cancel chan
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)
	defer stop()

	// "errgroup" instead of "WaitGroup" because:
	// - allows return an error from gorutine
	// - group errors from multiple gorutines into one
	// - wg.Add(1), wg.Done() - automatically under the hood, so never catch deadlock if you forget something ;-)
	// - allows orchestration of parallel processes through the context.Context(gracefull shut down)
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		a.logger.Info("starting "+a.cfg.App.Name, zap.String("addr", a.cfg.App.Host+":"+a.cfg.App.Port))
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server "+a.cfg.App.Name+" error: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		a.cache.BackupWorker(ctx)
		return nil
	})

	a.scorer.RunScorerPool(ctx, poolSize)

	g.Go(func() error {
		a.lbMemory.RunLBWorker(ctx)
		return nil
	})

	<-ctx.Done()

	a.scorer.ClosePool(ctx)

	a.logger.Info("shutting down " + a.cfg.App.Name + " gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if a.httpSrv != nil {
		if err := a.httpSrv.Shutdown(shutdownCtx); err != nil {
			a.logger.Error("http server shutdown "+a.cfg.App.Name+" error", zap.Error(err))
			return err
		}
	}

	if err := g.Wait(); err != nil {
		a.logger.Error(a.cfg.App.Name+" returning an error", zap.Error(err))
		return err
	}

	a.logger.Info(a.cfg.App.Name + " gracefully stopped")

	return nil
}

func (a *App) InitControllers(_ context.Context) {
	// services
	eventService := services.NewEventService(a.cache, a.scorer, a.metrics)
	lbService := services.NewLeaderboardService(a.lbMemory)

	// controllers
	rest.NewEventController(a.mux, eventService)
	rest.NewLeaderboardController(a.mux, lbService)

	// ops
	a.mux.HandleFunc(http.MethodGet+rest.Space+rest.RouteHealth, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	a.mux.HandleFunc(http.MethodGet+rest.Space+rest.RouteMetrics, func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
}

func (a *App) Logger() *zap.Logger { return a.logger }
