package metrics

import (
	"github.com/bastion-framework/bast"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	bast.Module
	Service *Service
}

func New(pool *pgxpool.Pool, sessionGuard, agentGuard bast.Guard, publisher metricPublisher) Module {
	r := newRepo(pool)
	s := newService(r, publisher)
	return Module{
		Module: bast.Module{
			Prefix:     "/metrics",
			Controller: newController(s, sessionGuard, agentGuard),
			Doc: bast.ModuleDoc{
				Name:        "Metrics",
				Description: "Time-series metric ingestion (agent) and query (user).",
			},
		},
		Service: s,
	}
}
