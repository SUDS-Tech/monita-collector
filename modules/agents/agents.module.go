package agents

import (
	"github.com/bastion-framework/bast"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module is Pattern B: exposes Service so AgentAuthGuard can be wired in main.
type Module struct {
	bast.Module
	Service *Service
}

func New(pool *pgxpool.Pool, sessionGuard bast.Guard) Module {
	r := newRepo(pool)
	s := newService(r)
	return Module{
		Module: bast.Module{
			Prefix:     "/agents",
			Guards:     []bast.Guard{sessionGuard},
			Controller: newController(s),
			Doc: bast.ModuleDoc{
				Name:        "Agents",
				Description: "Collector agents — creation, management, and token lifecycle.",
			},
		},
		Service: s,
	}
}