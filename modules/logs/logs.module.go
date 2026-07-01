package logs

import (
	"github.com/bastion-framework/bast"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool, sessionGuard, agentGuard bast.Guard) bast.Module {
	r := newRepo(pool)
	s := newService(r)
	return bast.Module{
		Prefix:     "/logs",
		Controller: newController(s, sessionGuard, agentGuard),
		Doc: bast.ModuleDoc{
			Name:        "Logs",
			Description: "Log entry ingestion (agent) and query/search (user).",
		},
	}
}
