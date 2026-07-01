package stream

import (
	"github.com/bastion-framework/bast"
	"github.com/goccy/go-json"
)

type controller struct {
	hub *Hub
}

func newController(hub *Hub) *controller {
	return &controller{hub: hub}
}

func (c *controller) Routes() []bast.Route {
	return []bast.Route{
		bast.STREAM("/metrics", c.StreamMetrics, bast.WithDoc(bast.Doc{
			Summary:     "Stream live metric events for an agent (SSE)",
			Description: "Emits 'metric' events as JSON whenever the agent ingests new data.",
			Tags:        []string{"Stream"},
			Params:      []bast.Param{bast.QueryParam("agent_id", "Agent UUID to subscribe to")},
		})),
		bast.STREAM("/logs", c.StreamLogs, bast.WithDoc(bast.Doc{
			Summary:     "Stream live log events for an agent (SSE)",
			Description: "Emits 'log' events as JSON whenever the agent ingests new entries.",
			Tags:        []string{"Stream"},
			Params:      []bast.Param{bast.QueryParam("agent_id", "Agent UUID to subscribe to")},
		})),
	}
}

func (c *controller) StreamMetrics(sctx *bast.StreamCtx) {
	agentID := sctx.Request.URL.Query().Get("agent_id")
	if agentID == "" {
		return
	}

	sctx.SetHeader("Content-Type", "text/event-stream")
	sctx.SetHeader("Cache-Control", "no-cache")
	sctx.SetHeader("Connection", "keep-alive")

	ch, unsub := c.hub.subscribeMetrics(agentID)
	defer unsub()

	for {
		select {
		case <-sctx.Closed():
			return
		case e, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(e)
			if err != nil {
				continue
			}
			if err := sctx.Send("metric", string(data)); err != nil {
				return
			}
			_ = sctx.Flush()
		}
	}
}

func (c *controller) StreamLogs(sctx *bast.StreamCtx) {
	agentID := sctx.Request.URL.Query().Get("agent_id")
	if agentID == "" {
		return
	}

	sctx.SetHeader("Content-Type", "text/event-stream")
	sctx.SetHeader("Cache-Control", "no-cache")
	sctx.SetHeader("Connection", "keep-alive")

	ch, unsub := c.hub.subscribeLogs(agentID)
	defer unsub()

	for {
		select {
		case <-sctx.Closed():
			return
		case e, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(e)
			if err != nil {
				continue
			}
			if err := sctx.Send("log", string(data)); err != nil {
				return
			}
			_ = sctx.Flush()
		}
	}
}
