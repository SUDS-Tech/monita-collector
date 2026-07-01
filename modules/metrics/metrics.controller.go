package metrics

import (
	"strconv"
	"time"

	"github.com/bastion-framework/bast"
	"github.com/google/uuid"

	apperr "github.com/SUDS-Tech/monita-collector/shared/errors"
	"github.com/SUDS-Tech/monita-collector/shared/guards"
)

type controller struct {
	svc          *Service
	sessionGuard bast.Guard
	agentGuard   bast.Guard
}

func newController(svc *Service, sessionGuard, agentGuard bast.Guard) *controller {
	return &controller{svc: svc, sessionGuard: sessionGuard, agentGuard: agentGuard}
}

func (c *controller) Routes() []bast.Route {
	return []bast.Route{
		bast.POST("/ingest", c.Ingest, bast.WithGuards(c.agentGuard), bast.WithDoc(bast.Doc{
			Summary:     "Ingest a batch of metric points",
			Description: "Called by the collector agent. Authenticated via agent token + HMAC.",
			Tags:        []string{"Metrics"},
			Body:        bast.Body[IngestRequest](),
		})),
		bast.GET("", c.Query, bast.WithGuards(c.sessionGuard), bast.WithDoc(bast.Doc{
			Summary: "Query metric points",
			Tags:    []string{"Metrics"},
			Params: []bast.Param{
				bast.QueryParam("agent_id", "Agent UUID"),
				bast.QueryParam("metric_name", "Metric name"),
				bast.QueryParam("from", "Start time (RFC3339)"),
				bast.QueryParam("to", "End time (RFC3339), defaults to now"),
				bast.QueryParam("limit", "Max results (default 1000, max 10000)"),
			},
			Returns: bast.Returns{200: bast.Body[[]PointResponse]()},
		})),
		bast.GET("/names", c.ListNames, bast.WithGuards(c.sessionGuard), bast.WithDoc(bast.Doc{
			Summary: "List distinct metric names for an agent",
			Tags:    []string{"Metrics"},
			Params:  []bast.Param{bast.QueryParam("agent_id", "Agent UUID")},
			Returns: bast.Returns{200: bast.Body[[]string]()},
		})),
	}
}

func (c *controller) Ingest(ctx *bast.Ctx) bast.Response {
	agent := ctx.MustGet("agent").(*guards.AgentInfo)
	agentID, err := uuid.Parse(agent.ID)
	if err != nil {
		return ctx.Error(bast.ErrUnauthorized(apperr.CodeAgentUnknown, "invalid agent ID"))
	}

	var req IngestRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Error(err)
	}
	if len(req.Points) == 0 {
		return ctx.NoContent()
	}

	n, err := c.svc.Ingest(ctx.Context(), agentID, req)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(map[string]int64{"ingested": n})
}

func (c *controller) Query(ctx *bast.Ctx) bast.Response {
	agentID, err := uuid.Parse(ctx.Query("agent_id"))
	if err != nil {
		return ctx.Error(bast.ErrBadRequest("INVALID_AGENT_ID", "agent_id must be a valid UUID"))
	}

	metricName := ctx.Query("metric_name")
	if metricName == "" {
		return ctx.Error(bast.ErrBadRequest("MISSING_PARAM", "metric_name is required"))
	}

	from, err := time.Parse(time.RFC3339, ctx.Query("from"))
	if err != nil {
		return ctx.Error(bast.ErrBadRequest("INVALID_FROM", "from must be RFC3339"))
	}

	to := time.Now()
	if raw := ctx.Query("to"); raw != "" {
		to, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return ctx.Error(bast.ErrBadRequest("INVALID_TO", "to must be RFC3339"))
		}
	}

	limit := int32(1000)
	if raw := ctx.Query("limit"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n < 1 || n > 10000 {
			return ctx.Error(bast.ErrBadRequest("INVALID_LIMIT", "limit must be 1-10000"))
		}
		limit = int32(n)
	}

	points, err := c.svc.Query(ctx.Context(), agentID, metricName, from, to, limit)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(points)
}

func (c *controller) ListNames(ctx *bast.Ctx) bast.Response {
	agentID, err := uuid.Parse(ctx.Query("agent_id"))
	if err != nil {
		return ctx.Error(bast.ErrBadRequest("INVALID_AGENT_ID", "agent_id must be a valid UUID"))
	}

	names, err := c.svc.ListNames(ctx.Context(), agentID)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(names)
}
