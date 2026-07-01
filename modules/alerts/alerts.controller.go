package alerts

import (
	"strconv"

	"github.com/bastion-framework/bast"
	"github.com/google/uuid"

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
		// Rules
		bast.POST("/rules", c.CreateRule, bast.WithDoc(bast.Doc{
			Summary: "Create an alert rule",
			Tags:    []string{"Alerts"},
			Body:    bast.Body[CreateRuleRequest](),
			Returns: bast.Returns{201: bast.Body[RuleResponse]()},
		})),
		bast.GET("/rules", c.ListRules, bast.WithDoc(bast.Doc{
			Summary: "List alert rules",
			Tags:    []string{"Alerts"},
			Returns: bast.Returns{200: bast.Body[[]RuleSummary]()},
		})),
		bast.GET("/rules/:id", c.GetRule, bast.WithDoc(bast.Doc{
			Summary: "Get an alert rule",
			Tags:    []string{"Alerts"},
			Params:  []bast.Param{bast.PathParam("id", "Rule UUID")},
			Returns: bast.Returns{200: bast.Body[RuleResponse]()},
		})),
		bast.PUT("/rules/:id", c.UpdateRule, bast.WithDoc(bast.Doc{
			Summary: "Update an alert rule",
			Tags:    []string{"Alerts"},
			Params:  []bast.Param{bast.PathParam("id", "Rule UUID")},
			Body:    bast.Body[UpdateRuleRequest](),
			Returns: bast.Returns{200: bast.Body[RuleResponse]()},
		})),
		bast.DELETE("/rules/:id", c.DeleteRule, bast.WithDoc(bast.Doc{
			Summary: "Delete an alert rule",
			Tags:    []string{"Alerts"},
			Params:  []bast.Param{bast.PathParam("id", "Rule UUID")},
		})),
		// Events
		bast.GET("/events", c.ListFiring, bast.WithDoc(bast.Doc{
			Summary: "List currently firing events for the org",
			Tags:    []string{"Alerts"},
			Params:  []bast.Param{bast.QueryParam("limit", "Max results (default 100)")},
			Returns: bast.Returns{200: bast.Body[[]FiringEventResponse]()},
		})),
		bast.GET("/rules/:id/events", c.ListEvents, bast.WithDoc(bast.Doc{
			Summary: "List events for a rule",
			Tags:    []string{"Alerts"},
			Params: []bast.Param{
				bast.PathParam("id", "Rule UUID"),
				bast.QueryParam("limit", "Max results (default 100)"),
			},
			Returns: bast.Returns{200: bast.Body[[]EventResponse]()},
		})),
		bast.POST("/events/:id/resolve", c.ResolveEvent, bast.WithDoc(bast.Doc{
			Summary: "Resolve a firing alert event",
			Tags:    []string{"Alerts"},
			Params:  []bast.Param{bast.PathParam("id", "Event UUID")},
		})),
	}
}

func (c *controller) CreateRule(ctx *bast.Ctx) bast.Response {
	orgID, err := orgFromClaims(ctx)
	if err != nil {
		return ctx.Error(err)
	}
	var req CreateRuleRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Error(err)
	}
	resp, err := c.svc.CreateRule(ctx.Context(), orgID, req)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.Created(resp)
}

func (c *controller) GetRule(ctx *bast.Ctx) bast.Response {
	orgID, err := orgFromClaims(ctx)
	if err != nil {
		return ctx.Error(err)
	}
	ruleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.Error(bast.ErrNotFound("RULE_NOT_FOUND", "alert rule not found"))
	}
	resp, err := c.svc.GetRule(ctx.Context(), orgID, ruleID)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(resp)
}

func (c *controller) ListRules(ctx *bast.Ctx) bast.Response {
	orgID, err := orgFromClaims(ctx)
	if err != nil {
		return ctx.Error(err)
	}
	rules, err := c.svc.ListRules(ctx.Context(), orgID)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(rules)
}

func (c *controller) UpdateRule(ctx *bast.Ctx) bast.Response {
	orgID, err := orgFromClaims(ctx)
	if err != nil {
		return ctx.Error(err)
	}
	ruleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.Error(bast.ErrNotFound("RULE_NOT_FOUND", "alert rule not found"))
	}
	var req UpdateRuleRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Error(err)
	}
	resp, err := c.svc.UpdateRule(ctx.Context(), orgID, ruleID, req)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(resp)
}

func (c *controller) DeleteRule(ctx *bast.Ctx) bast.Response {
	orgID, err := orgFromClaims(ctx)
	if err != nil {
		return ctx.Error(err)
	}
	ruleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.Error(bast.ErrNotFound("RULE_NOT_FOUND", "alert rule not found"))
	}
	if err := c.svc.DeleteRule(ctx.Context(), orgID, ruleID); err != nil {
		return ctx.Error(err)
	}
	return ctx.NoContent()
}

func (c *controller) ListFiring(ctx *bast.Ctx) bast.Response {
	orgID, err := orgFromClaims(ctx)
	if err != nil {
		return ctx.Error(err)
	}
	limit := int32(100)
	if raw := ctx.Query("limit"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n < 1 || n > 1000 {
			return ctx.Error(bast.ErrBadRequest("INVALID_LIMIT", "limit must be 1-1000"))
		}
		limit = int32(n)
	}
	events, err := c.svc.ListFiring(ctx.Context(), orgID, limit)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(events)
}

func (c *controller) ListEvents(ctx *bast.Ctx) bast.Response {
	ruleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.Error(bast.ErrNotFound("RULE_NOT_FOUND", "alert rule not found"))
	}
	limit := int32(100)
	if raw := ctx.Query("limit"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n < 1 || n > 1000 {
			return ctx.Error(bast.ErrBadRequest("INVALID_LIMIT", "limit must be 1-1000"))
		}
		limit = int32(n)
	}
	events, err := c.svc.ListEvents(ctx.Context(), ruleID, limit)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(events)
}

func (c *controller) ResolveEvent(ctx *bast.Ctx) bast.Response {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.Error(bast.ErrNotFound("EVENT_NOT_FOUND", "alert event not found"))
	}
	if err := c.svc.ResolveEvent(ctx.Context(), eventID); err != nil {
		return ctx.Error(err)
	}
	return ctx.NoContent()
}

func orgFromClaims(ctx *bast.Ctx) (uuid.UUID, error) {
	claims := ctx.MustGet("user").(*guards.UserClaims)
	id, err := uuid.Parse(claims.OrgID)
	if err != nil {
		return uuid.Nil, bast.ErrUnauthorized("INVALID_TOKEN", "invalid org ID in token")
	}
	return id, nil
}
