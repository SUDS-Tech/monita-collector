-- name: IngestMetricPoints :copyfrom
INSERT INTO metric_points (agent_id, metric_name, value, labels, recorded_at)
VALUES ($1, $2, $3, $4, $5);

-- name: QueryMetricPoints :many
SELECT agent_id, metric_name, value, labels, recorded_at, received_at
FROM metric_points
WHERE agent_id = $1
  AND metric_name = $2
  AND recorded_at >= $3
  AND recorded_at < $4
ORDER BY recorded_at DESC
LIMIT $5;

-- name: ListMetricNames :many
SELECT DISTINCT metric_name
FROM metric_points
WHERE agent_id = $1
ORDER BY metric_name;
