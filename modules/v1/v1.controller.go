package v1

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"time"

	"github.com/bastion-framework/bast"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/klauspost/compress/zstd"

	"github.com/SUDS-Tech/monita-collector/modules/logs"
	"github.com/SUDS-Tech/monita-collector/modules/metrics"
	"github.com/SUDS-Tech/monita-collector/shared/guards"
)

type metricIngestor interface {
	Ingest(ctx context.Context, agentID uuid.UUID, req metrics.IngestRequest) (int64, error)
}

type logIngestor interface {
	Ingest(ctx context.Context, agentID uuid.UUID, req logs.IngestRequest) (int64, error)
}

type fingerprintWriter interface {
	SetFingerprintHash(ctx context.Context, agentID, hash string) error
}

type tokenRotator interface {
	RotateToken(ctx context.Context, agentID string) (token, signingKey string, err error)
}

type controller struct {
	metrics     metricIngestor
	logs        logIngestor
	fingerprint fingerprintWriter
	rotation    tokenRotator
	agentGuard  bast.Guard
}

func newController(m metricIngestor, l logIngestor, fp fingerprintWriter, r tokenRotator, ag bast.Guard) *controller {
	return &controller{metrics: m, logs: l, fingerprint: fp, rotation: r, agentGuard: ag}
}

func (c *controller) Routes() []bast.Route {
	return []bast.Route{
		bast.POST("/metrics", c.IngestMetrics, bast.WithGuards(c.agentGuard), bast.WithDoc(bast.Doc{
			Summary:     "Ingest metric points (protocol v1)",
			Description: "zstd/gzip compressed JSON. See PROTOCOL.md §3.2.",
			Tags:        []string{"v1"},
		})),
		bast.POST("/logs", c.IngestLogs, bast.WithGuards(c.agentGuard), bast.WithDoc(bast.Doc{
			Summary:     "Ingest log entries (protocol v1)",
			Description: "zstd/gzip compressed JSON. See PROTOCOL.md §3.3.",
			Tags:        []string{"v1"},
		})),
		bast.POST("/agents/:id/fingerprint", c.RegisterFingerprint, bast.WithGuards(c.agentGuard), bast.WithDoc(bast.Doc{
			Summary:     "Register device fingerprint (protocol v1)",
			Description: "Called once after provisioning, before first push. See PROTOCOL.md §4.2.",
			Tags:        []string{"v1"},
		})),
		bast.POST("/agents/self/rotate", c.RotateToken, bast.WithGuards(c.agentGuard), bast.WithDoc(bast.Doc{
			Summary:     "Rotate agent token and signing key (protocol v1)",
			Description: "Called when rotation_required is true in an ingest response. See PROTOCOL.md §5.",
			Tags:        []string{"v1"},
		})),
	}
}

func (c *controller) IngestMetrics(ctx *bast.Ctx) bast.Response {
	agent := ctx.MustGet("agent").(*guards.AgentInfo)
	agentID, err := uuid.Parse(agent.ID)
	if err != nil {
		return ctx.Error(bast.ErrUnauthorized("AGENT_UNKNOWN", "invalid agent ID"))
	}

	var body metricsBody
	if err := decodeBody(ctx, &body); err != nil {
		return ctx.Error(err)
	}
	if len(body.Points) == 0 {
		return ctx.NoContent()
	}

	points := make([]metrics.IngestPoint, len(body.Points))
	for i, p := range body.Points {
		points[i] = metrics.IngestPoint{
			MetricName: p.Metric,
			Value:      p.Value,
			Labels:     p.Labels,
			RecordedAt: time.Unix(p.Ts, 0).UTC(),
		}
	}
	n, err := c.metrics.Ingest(ctx.Context(), agentID, metrics.IngestRequest{Points: points})
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(ingestResponse{Ingested: n, RotationRequired: agent.RotationRequired})
}

func (c *controller) IngestLogs(ctx *bast.Ctx) bast.Response {
	agent := ctx.MustGet("agent").(*guards.AgentInfo)
	agentID, err := uuid.Parse(agent.ID)
	if err != nil {
		return ctx.Error(bast.ErrUnauthorized("AGENT_UNKNOWN", "invalid agent ID"))
	}

	var body logsBody
	if err := decodeBody(ctx, &body); err != nil {
		return ctx.Error(err)
	}
	if len(body.Entries) == 0 {
		return ctx.NoContent()
	}

	entries := make([]logs.LogEntry, len(body.Entries))
	for i, e := range body.Entries {
		entries[i] = logs.LogEntry{
			Source:    e.Source,
			Level:     e.Level,
			Message:   e.Message,
			Count:     e.Count,
			FirstSeen: time.Unix(e.FirstSeen, 0).UTC(),
			LastSeen:  time.Unix(e.LastSeen, 0).UTC(),
		}
	}
	n, err := c.logs.Ingest(ctx.Context(), agentID, logs.IngestRequest{Entries: entries})
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(ingestResponse{Ingested: n, RotationRequired: agent.RotationRequired})
}

func (c *controller) RegisterFingerprint(ctx *bast.Ctx) bast.Response {
	agent := ctx.MustGet("agent").(*guards.AgentInfo)

	var body fingerprintBody
	if err := decodeBody(ctx, &body); err != nil {
		return ctx.Error(err)
	}
	if body.FingerprintHash == "" {
		return ctx.Error(bast.ErrBadRequest("MISSING_FINGERPRINT", "fingerprint_hash is required"))
	}

	if err := c.fingerprint.SetFingerprintHash(ctx.Context(), agent.ID, body.FingerprintHash); err != nil {
		return ctx.Error(err)
	}
	return ctx.NoContent()
}

func (c *controller) RotateToken(ctx *bast.Ctx) bast.Response {
	agent := ctx.MustGet("agent").(*guards.AgentInfo)

	token, signingKey, err := c.rotation.RotateToken(ctx.Context(), agent.ID)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(rotateResponse{Token: token, SigningKey: signingKey})
}

// decodeBody decompresses (zstd or gzip) and JSON-decodes the request body.
// The raw body was already buffered by the HMAC guard for signature verification.
func decodeBody(ctx *bast.Ctx, dst any) error {
	raw, err := ctx.RawBody()
	if err != nil {
		return bast.ErrBadRequest("INVALID_BODY", "cannot read body")
	}
	data, err := decompress(ctx.Header("Content-Encoding"), raw)
	if err != nil {
		return bast.ErrBadRequest("DECOMPRESS_ERROR", "cannot decompress body: "+err.Error())
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return bast.ErrBadRequest("INVALID_JSON", "invalid JSON body")
	}
	return nil
}

func decompress(encoding string, data []byte) ([]byte, error) {
	switch encoding {
	case "zstd":
		r, err := zstd.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer r.Close()
		return io.ReadAll(r)
	case "gzip":
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer r.Close()
		return io.ReadAll(r)
	default:
		return data, nil
	}
}
