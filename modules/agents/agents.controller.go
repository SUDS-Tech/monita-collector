package agents

import (
	"github.com/bastion-framework/bast"
	"github.com/google/uuid"

	apperr "github.com/SUDS-Tech/monita-collector/shared/errors"
	"github.com/SUDS-Tech/monita-collector/shared/guards"
)

type controller struct {
	svc *Service
}

func newController(svc *Service) *controller {
	return &controller{svc: svc}
}

func (c *controller) Routes() []bast.Route {
	return []bast.Route{
		bast.POST("", c.Create, bast.WithDoc(bast.Doc{
			Summary: "Create a new agent",
			Tags:    []string{"Agents"},
			Body:    bast.Body[CreateAgentRequest](),
			Returns: bast.Returns{201: bast.Body[CreateAgentResponse]()},
		})),
		bast.GET("", c.List, bast.WithDoc(bast.Doc{
			Summary: "List agents for the authenticated org",
			Tags:    []string{"Agents"},
			Returns: bast.Returns{200: bast.Body[[]AgentResponse]()},
		})),
		bast.GET("/:id", c.Get, bast.WithDoc(bast.Doc{
			Summary: "Get a single agent",
			Tags:    []string{"Agents"},
			Params:  []bast.Param{bast.PathParam("id", "Agent UUID")},
			Returns: bast.Returns{200: bast.Body[AgentResponse]()},
		})),
		bast.POST("/:id/revoke", c.Revoke, bast.WithDoc(bast.Doc{
			Summary: "Revoke an agent token",
			Tags:    []string{"Agents"},
			Params:  []bast.Param{bast.PathParam("id", "Agent UUID")},
		})),
		bast.DELETE("/:id", c.Delete, bast.WithDoc(bast.Doc{
			Summary: "Delete an agent",
			Tags:    []string{"Agents"},
			Params:  []bast.Param{bast.PathParam("id", "Agent UUID")},
		})),
	}
}

func (c *controller) Create(ctx *bast.Ctx) bast.Response {
	orgID, err := orgFromClaims(ctx)
	if err != nil {
		return ctx.Error(err)
	}
	var req CreateAgentRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Error(err)
	}
	resp, err := c.svc.Create(ctx.Context(), orgID, req)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.Created(resp)
}

func (c *controller) List(ctx *bast.Ctx) bast.Response {
	orgID, err := orgFromClaims(ctx)
	if err != nil {
		return ctx.Error(err)
	}
	agents, err := c.svc.List(ctx.Context(), orgID)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(agents)
}

func (c *controller) Get(ctx *bast.Ctx) bast.Response {
	orgID, err := orgFromClaims(ctx)
	if err != nil {
		return ctx.Error(err)
	}
	agentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.Error(bast.ErrNotFound("AGENT_NOT_FOUND", "agent not found"))
	}
	agent, err := c.svc.Get(ctx.Context(), orgID, agentID)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(agent)
}

func (c *controller) Revoke(ctx *bast.Ctx) bast.Response {
	orgID, err := orgFromClaims(ctx)
	if err != nil {
		return ctx.Error(err)
	}
	agentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.Error(bast.ErrNotFound("AGENT_NOT_FOUND", "agent not found"))
	}
	if err := c.svc.Revoke(ctx.Context(), orgID, agentID); err != nil {
		return ctx.Error(err)
	}
	return ctx.NoContent()
}

func (c *controller) Delete(ctx *bast.Ctx) bast.Response {
	orgID, err := orgFromClaims(ctx)
	if err != nil {
		return ctx.Error(err)
	}
	agentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.Error(bast.ErrNotFound("AGENT_NOT_FOUND", "agent not found"))
	}
	if err := c.svc.Delete(ctx.Context(), orgID, agentID); err != nil {
		return ctx.Error(err)
	}
	return ctx.NoContent()
}

func orgFromClaims(ctx *bast.Ctx) (uuid.UUID, error) {
	claims := ctx.MustGet("user").(*guards.UserClaims)
	id, err := uuid.Parse(claims.OrgID)
	if err != nil {
		return uuid.Nil, bast.ErrUnauthorized(apperr.CodeInvalidCredentials, "invalid org ID in token")
	}
	return id, nil
}