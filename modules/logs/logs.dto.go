package logs

import "time"

type LogEntry struct {
	Source    string    `json:"source"     validate:"required,min=1,max=200"`
	Level     string    `json:"level"      validate:"required,oneof=debug info warn error"`
	Message   string    `json:"message"    validate:"required,min=1"`
	Count     int32     `json:"count"      validate:"min=0"`
	FirstSeen time.Time `json:"first_seen" validate:"required"`
	LastSeen  time.Time `json:"last_seen"  validate:"required"`
}

type IngestRequest struct {
	Entries []LogEntry `json:"entries" validate:"required,min=1,max=1000,dive"`
}

type EntryResponse struct {
	AgentID    string    `json:"agent_id"`
	Source     string    `json:"source"`
	Level      string    `json:"level"`
	Message    string    `json:"message"`
	Count      int32     `json:"count"`
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
	ReceivedAt time.Time `json:"received_at"`
}
