package logs

import (
	"github.com/bastion-framework/bast"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	bast.Module
	Service *Service
}

func New(pool *pgxpool.Pool, sessionGuard, agentGuard bast.Guard, publisher logPublisher) Module {
	r := newRepo(pool)
	s := newService(r, publisher)
	return Module{
		Module: bast.Module{
			Prefix:     "/logs",
			Controller: newController(s, sessionGuard, agentGuard),
			Doc: bast.ModuleDoc{
				Name:        "Logs",
				Description: "Log entry ingestion (agent) and query/search (user).",
			},
		},
		Service: s,
	}
}
