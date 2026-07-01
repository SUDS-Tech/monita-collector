package logs

import "time"

type LogEntry struct {
	Source    string    `json:"source"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Count     int32     `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

type IngestRequest struct {
	Entries []LogEntry `json:"entries"`
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
