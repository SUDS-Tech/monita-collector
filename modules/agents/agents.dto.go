package agents

import "time"

type CreateAgentRequest struct {
	Name     string            `json:"name"`
	Hostname string            `json:"hostname"`
	Tags     map[string]string `json:"tags"`
}

type AgentResponse struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Hostname         string            `json:"hostname"`
	Tags             map[string]string `json:"tags"`
	Frozen           bool              `json:"frozen"`
	Revoked          bool              `json:"revoked"`
	RotationRequired bool              `json:"rotation_required"`
	ExpiresAt        time.Time         `json:"expires_at"`
	LastSeenAt       *time.Time        `json:"last_seen_at"`
	CreatedAt        time.Time         `json:"created_at"`
}

// CreateAgentResponse is returned once on creation — Token and SigningKey are never shown again.
type CreateAgentResponse struct {
	AgentResponse
	Token      string `json:"token"`
	SigningKey string `json:"signing_key"`
}
