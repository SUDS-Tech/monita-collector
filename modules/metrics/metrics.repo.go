package metrics

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	dbsqlc "github.com/SUDS-Tech/monita-collector/internal/db/sqlc"
)

type repo struct {
	q *dbsqlc.Queries
}

func newRepo(pool *pgxpool.Pool) *repo {
	return &repo{q: dbsqlc.New(pool)}
}

func (r *repo) ingest(ctx context.Context, rows []dbsqlc.IngestMetricPointsParams) (int64, error) {
	return r.q.IngestMetricPoints(ctx, rows)
}

func (r *repo) query(ctx context.Context, agentID uuid.UUID, metricName string, from, to time.Time, limit int32) ([]dbsqlc.MetricPoint, error) {
	return r.q.QueryMetricPoints(ctx, dbsqlc.QueryMetricPointsParams{
		AgentID:      agentID,
		MetricName:   metricName,
		RecordedAt:   from,
		RecordedAt_2: to,
		Limit:        limit,
	})
}

func (r *repo) listNames(ctx context.Context, agentID uuid.UUID) ([]string, error) {
	return r.q.ListMetricNames(ctx, agentID)
}
