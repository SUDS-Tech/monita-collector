CREATE TABLE metric_points (
    agent_id    UUID             NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    metric_name TEXT             NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    labels      JSONB            NOT NULL DEFAULT '{}',
    recorded_at TIMESTAMPTZ      NOT NULL,
    received_at TIMESTAMPTZ      NOT NULL DEFAULT now()
) PARTITION BY RANGE (recorded_at);

CREATE TABLE metric_points_default PARTITION OF metric_points DEFAULT;

CREATE INDEX idx_metric_points_brin         ON metric_points USING BRIN (recorded_at);
CREATE INDEX idx_metric_points_agent_metric ON metric_points (agent_id, metric_name, recorded_at DESC);
