package users

import (
	"github.com/bastion-framework/bast"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module is Pattern B: embeds bast.Module and exposes Service
// so billing can call SetSubscription / FindOrgByStripeCustomer.
type Module struct {
	bast.Module
	Service *Service
}

func New(pool *pgxpool.Pool, jwtSecret string, sessionGuard bast.Guard) Module {
	r := newRepo(pool)
	s := newService(r, []byte(jwtSecret))
	return Module{
		Module: bast.Module{
			Prefix:     "/users",
			Controller: newController(s, sessionGuard),
			Doc: bast.ModuleDoc{
				Name:        "Users",
				Description: "Registration, login, and profile for authenticated users.",
			},
		},
		Service: s,
	}
}