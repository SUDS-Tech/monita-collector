package main

import (
	"context"
	"log"
	"time"

	"github.com/bastion-framework/bast"
	"github.com/bastion-framework/bast/middleware"
	"github.com/joho/godotenv"

	"github.com/SUDS-Tech/monita-collector/internal/config"
	"github.com/SUDS-Tech/monita-collector/internal/db"
	"github.com/SUDS-Tech/monita-collector/modules/agents"
	"github.com/SUDS-Tech/monita-collector/modules/logs"
	"github.com/SUDS-Tech/monita-collector/modules/metrics"
	"github.com/SUDS-Tech/monita-collector/modules/users"
	"github.com/SUDS-Tech/monita-collector/shared/guards"
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
	)

	metricsMod := metrics.New(pool, sessionAuth, agentAuth)

	logsMod := logs.New(pool, sessionAuth, agentAuth)

	app.Register(usersMod.Module)
	app.Register(agentsMod.Module)
	app.Register(metricsMod)
	app.Register(logsMod)

	app.Listen()
}
