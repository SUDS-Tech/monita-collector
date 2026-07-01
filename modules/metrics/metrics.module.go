package metrics

import (
	"github.com/bastion-framework/bast"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool, sessionGuard, agentGuard bast.Guard, publisher metricPublisher) bast.Module {
	r := newRepo(pool)
	s := newService(r, publisher)
	return bast.Module{
		Prefix:     "/metrics",
		Controller: newController(s, sessionGuard, agentGuard),
		Doc: bast.ModuleDoc{
			Name:        "Metrics",
			Description: "Time-series metric ingestion (agent) and query (user).",
		},
	}
}
