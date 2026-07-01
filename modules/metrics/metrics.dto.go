package metrics

import "time"

type IngestPoint struct {
	MetricName string            `json:"metric_name" validate:"required,min=1,max=200"`
	Value      float64           `json:"value"`
	Labels     map[string]string `json:"labels"`
	RecordedAt time.Time         `json:"recorded_at" validate:"required"`
}

type IngestRequest struct {
	Points []IngestPoint `json:"points" validate:"required,min=1,max=1000,dive"`
}

type PointResponse struct {
	AgentID    string            `json:"agent_id"`
	MetricName string            `json:"metric_name"`
	Value      float64           `json:"value"`
	Labels     map[string]string `json:"labels"`
	RecordedAt time.Time         `json:"recorded_at"`
	ReceivedAt time.Time         `json:"received_at"`
}
