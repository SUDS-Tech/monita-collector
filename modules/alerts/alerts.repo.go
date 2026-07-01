package alerts

import (
	"context"

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

func (r *repo) createRule(ctx context.Context, arg dbsqlc.CreateAlertRuleParams) (dbsqlc.CreateAlertRuleRow, error) {
	return r.q.CreateAlertRule(ctx, arg)
}

func (r *repo) getRule(ctx context.Context, id, orgID uuid.UUID) (dbsqlc.GetAlertRuleRow, error) {
	return r.q.GetAlertRule(ctx, dbsqlc.GetAlertRuleParams{ID: id, OrgID: orgID})
}

func (r *repo) listRules(ctx context.Context, orgID uuid.UUID) ([]dbsqlc.ListAlertRulesRow, error) {
	return r.q.ListAlertRules(ctx, orgID)
}

func (r *repo) updateRule(ctx context.Context, arg dbsqlc.UpdateAlertRuleParams) (dbsqlc.UpdateAlertRuleRow, error) {
	return r.q.UpdateAlertRule(ctx, arg)
}

func (r *repo) deleteRule(ctx context.Context, id, orgID uuid.UUID) error {
	return r.q.DeleteAlertRule(ctx, dbsqlc.DeleteAlertRuleParams{ID: id, OrgID: orgID})
}

func (r *repo) listEvents(ctx context.Context, ruleID uuid.UUID, limit int32) ([]dbsqlc.AlertEvent, error) {
	return r.q.ListAlertEvents(ctx, dbsqlc.ListAlertEventsParams{RuleID: ruleID, Limit: limit})
}

func (r *repo) listFiring(ctx context.Context, orgID uuid.UUID, limit int32) ([]dbsqlc.ListFiringEventsRow, error) {
	return r.q.ListFiringEvents(ctx, dbsqlc.ListFiringEventsParams{OrgID: orgID, Limit: limit})
}

func (r *repo) resolveEvent(ctx context.Context, id uuid.UUID) error {
	return r.q.ResolveAlertEvent(ctx, id)
}
