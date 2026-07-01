package metrics

import (
	"context"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"

	dbsqlc "github.com/SUDS-Tech/monita-collector/internal/db/sqlc"
)

type metricPublisher interface {
	PublishMetric(agentID, metricName string, value float64, labels map[string]string, recordedAt time.Time)
}

type Service struct {
	repo      *repo
	publisher metricPublisher
}

func newService(r *repo, publisher metricPublisher) *Service {
	return &Service{repo: r, publisher: publisher}
}

func (s *Service) Ingest(ctx context.Context, agentID uuid.UUID, req IngestRequest) (int64, error) {
	rows := make([]dbsqlc.IngestMetricPointsParams, len(req.Points))
	for i, p := range req.Points {
		labelsJSON, err := json.Marshal(p.Labels)
		if err != nil {
			return 0, err
		}
		rows[i] = dbsqlc.IngestMetricPointsParams{
			AgentID:    agentID,
			MetricName: p.MetricName,
			Value:      p.Value,
			Labels:     labelsJSON,
			RecordedAt: p.RecordedAt,
		}
	}
	n, err := s.repo.ingest(ctx, rows)
	if err != nil {
		return 0, err
	}
	if s.publisher != nil {
		for _, p := range req.Points {
			s.publisher.PublishMetric(agentID.String(), p.MetricName, p.Value, p.Labels, p.RecordedAt)
		}
	}
	return n, nil
}

func (s *Service) Query(ctx context.Context, agentID uuid.UUID, metricName string, from, to time.Time, limit int32) ([]PointResponse, error) {
	rows, err := s.repo.query(ctx, agentID, metricName, from, to, limit)
	if err != nil {
		return nil, err
	}
	out := make([]PointResponse, len(rows))
	for i, r := range rows {
		out[i] = toPointResponse(r)
	}
	return out, nil
}

func (s *Service) ListNames(ctx context.Context, agentID uuid.UUID) ([]string, error) {
	return s.repo.listNames(ctx, agentID)
}

func toPointResponse(r dbsqlc.MetricPoint) PointResponse {
	var labels map[string]string
	if len(r.Labels) > 0 {
		_ = json.Unmarshal(r.Labels, &labels)
	}
	if labels == nil {
		labels = map[string]string{}
	}
	return PointResponse{
		AgentID:    r.AgentID.String(),
		MetricName: r.MetricName,
		Value:      r.Value,
		Labels:     labels,
		RecordedAt: r.RecordedAt,
		ReceivedAt: r.ReceivedAt,
	}
}
