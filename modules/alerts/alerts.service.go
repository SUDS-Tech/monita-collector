package alerts

import (
	"context"
	"time"

	"github.com/bastion-framework/bast"
	"github.com/goccy/go-json"
	"github.com/google/uuid"

	dbsqlc "github.com/SUDS-Tech/monita-collector/internal/db/sqlc"
)

type Service struct {
	repo *repo
}

func newService(r *repo) *Service {
	return &Service{repo: r}
}

func (s *Service) CreateRule(ctx context.Context, orgID uuid.UUID, req CreateRuleRequest) (*RuleResponse, error) {
	cond, err := json.Marshal(req.Condition)
	if err != nil {
		return nil, err
	}
	channels, err := json.Marshal(req.NotificationChannels)
	if err != nil {
		return nil, err
	}

	row, err := s.repo.createRule(ctx, dbsqlc.CreateAlertRuleParams{
		OrgID:                orgID,
		Name:                 req.Name,
		Target:               dbsqlc.AlertTarget(req.Target),
		Condition:            cond,
		Severity:             dbsqlc.AlertSeverity(req.Severity),
		NotificationChannels: channels,
	})
	if err != nil {
		return nil, err
	}
	resp := toRuleResponse(row.ID, row.OrgID, row.Name, row.Target, row.Condition, row.Severity, row.NotificationChannels, row.Enabled, row.CreatedAt, row.UpdatedAt)
	return &resp, nil
}

func (s *Service) GetRule(ctx context.Context, orgID, ruleID uuid.UUID) (*RuleResponse, error) {
	row, err := s.repo.getRule(ctx, ruleID, orgID)
	if err != nil {
		return nil, bast.ErrNotFound("RULE_NOT_FOUND", "alert rule not found")
	}
	resp := toRuleResponse(row.ID, row.OrgID, row.Name, row.Target, row.Condition, row.Severity, row.NotificationChannels, row.Enabled, row.CreatedAt, row.UpdatedAt)
	return &resp, nil
}

func (s *Service) ListRules(ctx context.Context, orgID uuid.UUID) ([]RuleSummary, error) {
	rows, err := s.repo.listRules(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := make([]RuleSummary, len(rows))
	for i, r := range rows {
		out[i] = RuleSummary{
			ID:        r.ID.String(),
			Name:      r.Name,
			Target:    string(r.Target),
			Severity:  string(r.Severity),
			Enabled:   r.Enabled,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
	}
	return out, nil
}

func (s *Service) UpdateRule(ctx context.Context, orgID, ruleID uuid.UUID, req UpdateRuleRequest) (*RuleResponse, error) {
	cond, err := json.Marshal(req.Condition)
	if err != nil {
		return nil, err
	}
	channels, err := json.Marshal(req.NotificationChannels)
	if err != nil {
		return nil, err
	}

	row, err := s.repo.updateRule(ctx, dbsqlc.UpdateAlertRuleParams{
		ID:                   ruleID,
		OrgID:                orgID,
		Name:                 req.Name,
		Condition:            cond,
		Severity:             dbsqlc.AlertSeverity(req.Severity),
		NotificationChannels: channels,
		Enabled:              req.Enabled,
	})
	if err != nil {
		return nil, bast.ErrNotFound("RULE_NOT_FOUND", "alert rule not found")
	}
	resp := toRuleResponse(row.ID, row.OrgID, row.Name, row.Target, row.Condition, row.Severity, row.NotificationChannels, row.Enabled, row.CreatedAt, row.UpdatedAt)
	return &resp, nil
}

func (s *Service) DeleteRule(ctx context.Context, orgID, ruleID uuid.UUID) error {
	return s.repo.deleteRule(ctx, ruleID, orgID)
}

func (s *Service) ListEvents(ctx context.Context, ruleID uuid.UUID, limit int32) ([]EventResponse, error) {
	rows, err := s.repo.listEvents(ctx, ruleID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]EventResponse, len(rows))
	for i, r := range rows {
		var val map[string]any
		if len(r.ValueAtFire) > 0 {
			_ = json.Unmarshal(r.ValueAtFire, &val)
		}
		out[i] = EventResponse{
			ID:          r.ID.String(),
			RuleID:      r.RuleID.String(),
			AgentID:     r.AgentID.String(),
			FiredAt:     r.FiredAt,
			ResolvedAt:  derefTime(r.ResolvedAt),
			ValueAtFire: val,
			Status:      string(r.Status),
		}
	}
	return out, nil
}

func (s *Service) ListFiring(ctx context.Context, orgID uuid.UUID, limit int32) ([]FiringEventResponse, error) {
	rows, err := s.repo.listFiring(ctx, orgID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]FiringEventResponse, len(rows))
	for i, r := range rows {
		out[i] = FiringEventResponse{
			ID:      r.ID.String(),
			RuleID:  r.RuleID.String(),
			AgentID: r.AgentID.String(),
			FiredAt: r.FiredAt,
			Status:  string(r.Status),
		}
	}
	return out, nil
}

func (s *Service) ResolveEvent(ctx context.Context, eventID uuid.UUID) error {
	return s.repo.resolveEvent(ctx, eventID)
}

func toRuleResponse(id, orgID uuid.UUID, name string, target dbsqlc.AlertTarget, condition []byte, severity dbsqlc.AlertSeverity, channels []byte, enabled bool, createdAt, updatedAt time.Time) RuleResponse {
	var cond map[string]any
	_ = json.Unmarshal(condition, &cond)
	var notifChannels []map[string]any
	_ = json.Unmarshal(channels, &notifChannels)
	return RuleResponse{
		ID:                   id.String(),
		OrgID:                orgID.String(),
		Name:                 name,
		Target:               string(target),
		Condition:            cond,
		Severity:             string(severity),
		NotificationChannels: notifChannels,
		Enabled:              enabled,
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
	}
}

func derefTime(t **time.Time) *time.Time {
	if t == nil || *t == nil {
		return nil
	}
	return *t
}
