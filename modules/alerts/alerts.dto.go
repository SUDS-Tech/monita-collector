package alerts

import "time"

type CreateRuleRequest struct {
	Name                 string           `json:"name"`
	Target               string           `json:"target"`
	Condition            map[string]any   `json:"condition"`
	Severity             string           `json:"severity"`
	NotificationChannels []map[string]any `json:"notification_channels"`
}

type UpdateRuleRequest struct {
	Name                 string           `json:"name"`
	Condition            map[string]any   `json:"condition"`
	Severity             string           `json:"severity"`
	NotificationChannels []map[string]any `json:"notification_channels"`
	Enabled              bool             `json:"enabled"`
}

type RuleResponse struct {
	ID                   string           `json:"id"`
	OrgID                string           `json:"org_id"`
	Name                 string           `json:"name"`
	Target               string           `json:"target"`
	Condition            map[string]any   `json:"condition"`
	Severity             string           `json:"severity"`
	NotificationChannels []map[string]any `json:"notification_channels"`
	Enabled              bool             `json:"enabled"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

type RuleSummary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Target    string    `json:"target"`
	Severity  string    `json:"severity"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EventResponse struct {
	ID          string         `json:"id"`
	RuleID      string         `json:"rule_id"`
	AgentID     string         `json:"agent_id"`
	FiredAt     time.Time      `json:"fired_at"`
	ResolvedAt  *time.Time     `json:"resolved_at"`
	ValueAtFire map[string]any `json:"value_at_fire"`
	Status      string         `json:"status"`
}

type FiringEventResponse struct {
	ID      string    `json:"id"`
	RuleID  string    `json:"rule_id"`
	AgentID string    `json:"agent_id"`
	FiredAt time.Time `json:"fired_at"`
	Status  string    `json:"status"`
}
