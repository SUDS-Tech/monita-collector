package logs

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
			Summary:     "Ingest a batch of log entries",
			Description: "Called by the collector agent. Authenticated via agent token + HMAC.",
			Tags:        []string{"Logs"},
			Body:        bast.Body[IngestRequest](),
		})),
		bast.GET("", c.Query, bast.WithGuards(c.sessionGuard), bast.WithDoc(bast.Doc{
			Summary: "Query log entries",
			Tags:    []string{"Logs"},
			Params: []bast.Param{
				bast.QueryParam("agent_id", "Agent UUID"),
				bast.QueryParam("from", "Start time (RFC3339)"),
				bast.QueryParam("to", "End time (RFC3339), defaults to now"),
				bast.QueryParam("level", "Filter by level: debug, info, warn, error (optional)"),
				bast.QueryParam("limit", "Max results (default 1000, max 10000)"),
			},
			Returns: bast.Returns{200: bast.Body[[]EntryResponse]()},
		})),
		bast.GET("/search", c.Search, bast.WithGuards(c.sessionGuard), bast.WithDoc(bast.Doc{
			Summary: "Full-text search log messages",
			Tags:    []string{"Logs"},
			Params: []bast.Param{
				bast.QueryParam("agent_id", "Agent UUID"),
				bast.QueryParam("q", "Search query"),
				bast.QueryParam("from", "Start time (RFC3339)"),
				bast.QueryParam("to", "End time (RFC3339), defaults to now"),
				bast.QueryParam("limit", "Max results (default 100, max 1000)"),
			},
			Returns: bast.Returns{200: bast.Body[[]EntryResponse]()},
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
	if len(req.Entries) == 0 {
		return ctx.NoContent()
	}

	n, err := c.svc.Ingest(ctx.Context(), agentID, req)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(map[string]int64{"ingested": n})
}

func (c *controller) Query(ctx *bast.Ctx) bast.Response {
	agentID, from, to, limit, err := parseCommonParams(ctx, 1000, 10000)
	if err != nil {
		return ctx.Error(err)
	}

	var level *string
	if raw := ctx.Query("level"); raw != "" {
		level = &raw
	}

	entries, err := c.svc.Query(ctx.Context(), agentID, from, to, level, limit)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(entries)
}

func (c *controller) Search(ctx *bast.Ctx) bast.Response {
	agentID, from, to, limit, err := parseCommonParams(ctx, 100, 1000)
	if err != nil {
		return ctx.Error(err)
	}

	q := ctx.Query("q")
	if q == "" {
		return ctx.Error(bast.ErrBadRequest("MISSING_PARAM", "q is required"))
	}

	entries, err := c.svc.Search(ctx.Context(), agentID, from, to, q, limit)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(entries)
}

func parseCommonParams(ctx *bast.Ctx, defaultLimit, maxLimit int64) (uuid.UUID, time.Time, time.Time, int32, error) {
	agentID, err := uuid.Parse(ctx.Query("agent_id"))
	if err != nil {
		return uuid.Nil, time.Time{}, time.Time{}, 0, bast.ErrBadRequest("INVALID_AGENT_ID", "agent_id must be a valid UUID")
	}

	from, err := time.Parse(time.RFC3339, ctx.Query("from"))
	if err != nil {
		return uuid.Nil, time.Time{}, time.Time{}, 0, bast.ErrBadRequest("INVALID_FROM", "from must be RFC3339")
	}

	to := time.Now()
	if raw := ctx.Query("to"); raw != "" {
		to, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return uuid.Nil, time.Time{}, time.Time{}, 0, bast.ErrBadRequest("INVALID_TO", "to must be RFC3339")
		}
	}

	limit := defaultLimit
	if raw := ctx.Query("limit"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n < 1 || n > maxLimit {
			return uuid.Nil, time.Time{}, time.Time{}, 0, bast.ErrBadRequest("INVALID_LIMIT", "limit out of range")
		}
		limit = n
	}

	return agentID, from, to, int32(limit), nil
}
