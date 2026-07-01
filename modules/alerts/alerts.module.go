package alerts

import (
	"github.com/bastion-framework/bast"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool, sessionGuard bast.Guard) bast.Module {
	r := newRepo(pool)
	s := newService(r)
	return bast.Module{
		Prefix:     "/alerts",
		Guards:     []bast.Guard{sessionGuard},
		Controller: newController(s),
		Doc: bast.ModuleDoc{
			Name:        "Alerts",
			Description: "Alert rule management and event tracking.",
		},
	}
}
