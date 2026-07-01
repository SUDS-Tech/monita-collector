package stream

import "github.com/bastion-framework/bast"

// Module is Pattern B: exposes Hub so metrics and logs can publish to it.
type Module struct {
	bast.Module
	Hub *Hub
}

func New(sessionGuard bast.Guard) Module {
	hub := NewHub()
	return Module{
		Module: bast.Module{
			Prefix:  "/stream",
			Guards:  []bast.Guard{sessionGuard},
			Controller: newController(hub),
			Doc: bast.ModuleDoc{
				Name:        "Stream",
				Description: "Real-time Server-Sent Events for live metric and log data.",
			},
		},
		Hub: hub,
	}
}
