package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/bastion-framework/bast"
	"github.com/bastion-framework/bast/middleware"
	"github.com/joho/godotenv"

	"github.com/SUDS-Tech/monita-collector/internal/config"
	"github.com/SUDS-Tech/monita-collector/internal/db"
	"github.com/SUDS-Tech/monita-collector/modules/agents"
	"github.com/SUDS-Tech/monita-collector/modules/alerts"
	"github.com/SUDS-Tech/monita-collector/modules/logs"
	"github.com/SUDS-Tech/monita-collector/modules/metrics"
	"github.com/SUDS-Tech/monita-collector/modules/stream"
	"github.com/SUDS-Tech/monita-collector/modules/users"
	v1mod "github.com/SUDS-Tech/monita-collector/modules/v1"
	"github.com/SUDS-Tech/monita-collector/shared/guards"
	appMiddleware "github.com/SUDS-Tech/monita-collector/shared/middleware"
	"github.com/SUDS-Tech/monita-collector/shared/validate"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	sessionAuth := guards.NewSessionAuth(cfg.JWT.Secret)
	nonceCache := guards.NewMemoryNonceCache()

	usersMod := users.New(pool, cfg.JWT.Secret, sessionAuth)
	agentsMod := agents.New(pool, sessionAuth)
	agentAuth := guards.NewAgentAuth(agentsMod.Service, nonceCache)

	app := bast.New(bast.Config{
		Port:         cfg.Server.Port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		Validator:    validate.New(),
		Docs: &bast.DocsConfig{
			Enabled:     true,
			Path:        "/docs",
			JSONPath:    "/openapi.json",
			Title:       "Monita Collector API",
			Version:     "0.1.0",
			Description: "Open-source monitoring collector — AGPL v3.",
		},
		Health: &bast.HealthConfig{
			LivePath:  "/health",
			ReadyPath: "/ready",
			Checks: []bast.HealthCheck{
				bast.CustomCheck("postgres", func(ctx context.Context) error {
					return pool.Ping(ctx)
				}),
			},
		},
	})

	app.Use(
		middleware.RequestID,
		middleware.Recover,
		middleware.Logger,
		appMiddleware.RateLimit(100, 200),
	)

	streamMod := stream.New(sessionAuth)
	metricsMod := metrics.New(pool, sessionAuth, agentAuth, streamMod.Hub)
	logsSvc := logs.New(pool, sessionAuth, agentAuth, streamMod.Hub)
	alertsMod := alerts.New(pool, sessionAuth)

	app.Register(usersMod.Module)
	app.Register(agentsMod.Module)
	app.Register(metricsMod.Module)
	app.Register(logsSvc.Module)
	app.Register(alertsMod)
	app.Register(streamMod.Module)
	app.Register(v1mod.New(metricsMod.Service, logsSvc.Service, agentsMod.Service, agentsMod.Service, agentAuth))

	// Graceful shutdown on SIGTERM / SIGINT.
	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	errc := make(chan error, 1)
	go func() { errc <- app.Listen() }()

	select {
	case err := <-errc:
		if err != nil {
			log.Fatalf("server: %v", err)
		}
	case <-sigCtx.Done():
		stop()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := app.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
		<-errc
	}
}
