package v1

// Protocol §3.2 — Metrics push body (pre-compression JSON)
type metricsBody struct {
	AgentID string   `json:"agent_id"` // ignored — identity comes from auth token
	Points  []point  `json:"points"`
}

type point struct {
	Metric string            `json:"metric"`
	Value  float64           `json:"value"`
	Labels map[string]string `json:"labels"`
	Ts     int64             `json:"ts"` // unix seconds, agent clock
}

// Protocol §3.3 — Logs push body (pre-compression JSON)
type logsBody struct {
	AgentID string  `json:"agent_id"` // ignored — identity comes from auth token
	Entries []entry `json:"entries"`
}

type entry struct {
	Source    string `json:"source"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Count     int32  `json:"count"`
	FirstSeen int64  `json:"first_seen"` // unix seconds
	LastSeen  int64  `json:"last_seen"`  // unix seconds
}

// Protocol §3.4 — Fingerprint registration body
type fingerprintBody struct {
	AgentID         string          `json:"agent_id"`
	FingerprintHash string          `json:"fingerprint_hash"`
	Components      map[string]bool `json:"components"`
}

type ingestResponse struct {
	Ingested         int64 `json:"ingested"`
	RotationRequired bool  `json:"rotation_required"`
}

type rotateResponse struct {
	Token      string `json:"token"`
	SigningKey string `json:"signing_key"`
}
