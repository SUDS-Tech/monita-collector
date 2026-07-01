package agents

import (
	"context"
	"time"

	"github.com/goccy/go-json"

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

func (r *repo) createAgent(ctx context.Context, arg dbsqlc.CreateAgentParams) (dbsqlc.CreateAgentRow, error) {
	return r.q.CreateAgent(ctx, arg)
}

func (r *repo) getAgentByID(ctx context.Context, id, orgID uuid.UUID) (dbsqlc.GetAgentByIDRow, error) {
	return r.q.GetAgentByID(ctx, dbsqlc.GetAgentByIDParams{ID: id, OrgID: orgID})
}

func (r *repo) listAgentsByOrgID(ctx context.Context, orgID uuid.UUID) ([]dbsqlc.ListAgentsByOrgIDRow, error) {
	return r.q.ListAgentsByOrgID(ctx, orgID)
}

func (r *repo) getAgentByTokenHash(ctx context.Context, tokenHash string) (dbsqlc.GetAgentByTokenHashRow, error) {
	return r.q.GetAgentByTokenHash(ctx, tokenHash)
}

func (r *repo) updateLastSeen(ctx context.Context, id uuid.UUID) error {
	return r.q.UpdateAgentLastSeen(ctx, id)
}

func (r *repo) freezeAgent(ctx context.Context, id uuid.UUID) error {
	return r.q.FreezeAgent(ctx, id)
}

func (r *repo) setFingerprintHash(ctx context.Context, id uuid.UUID, hash string) error {
	return r.q.SetFingerprintHash(ctx, dbsqlc.SetFingerprintHashParams{ID: id, FingerprintHash: &hash})
}

func (r *repo) setFingerprintDrift(ctx context.Context, id uuid.UUID) error {
	return r.q.SetFingerprintDrift(ctx, id)
}

func (r *repo) revokeAgent(ctx context.Context, id, orgID uuid.UUID) error {
	return r.q.RevokeAgent(ctx, dbsqlc.RevokeAgentParams{ID: id, OrgID: orgID})
}

func (r *repo) deleteAgent(ctx context.Context, id, orgID uuid.UUID) error {
	return r.q.DeleteAgent(ctx, dbsqlc.DeleteAgentParams{ID: id, OrgID: orgID})
}

func tagsToMap(raw []byte) map[string]string {
	m := map[string]string{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &m)
	}
	return m
}

func derefLastSeen(t **time.Time) *time.Time {
	if t == nil || *t == nil {
		return nil
	}
	return *t
}
