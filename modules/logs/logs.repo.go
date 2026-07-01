package logs

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

func (r *repo) ingest(ctx context.Context, rows []dbsqlc.IngestLogEntriesParams) (int64, error) {
	return r.q.IngestLogEntries(ctx, rows)
}

func (r *repo) query(ctx context.Context, agentID uuid.UUID, from, to time.Time, level *string, limit int32) ([]dbsqlc.QueryLogEntriesRow, error) {
	return r.q.QueryLogEntries(ctx, dbsqlc.QueryLogEntriesParams{
		AgentID:      agentID,
		ReceivedAt:   from,
		ReceivedAt_2: to,
		Level:        level,
		Limit:        limit,
	})
}

func (r *repo) search(ctx context.Context, agentID uuid.UUID, from, to time.Time, query string, limit int32) ([]dbsqlc.SearchLogEntriesRow, error) {
	return r.q.SearchLogEntries(ctx, dbsqlc.SearchLogEntriesParams{
		AgentID:        agentID,
		ReceivedAt:     from,
		ReceivedAt_2:   to,
		PlaintoTsquery: query,
		Limit:          limit,
	})
}
