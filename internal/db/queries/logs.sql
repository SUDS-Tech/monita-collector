-- name: IngestLogEntries :copyfrom
INSERT INTO log_entries (agent_id, source, level, message, count, first_seen, last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: QueryLogEntries :many
SELECT agent_id, source, level, message, count, first_seen, last_seen, received_at
FROM log_entries
WHERE agent_id = $1
  AND received_at >= $2
  AND received_at < $3
  AND (sqlc.narg('level')::text IS NULL OR level = sqlc.narg('level'))
ORDER BY received_at DESC
LIMIT $4;

-- name: SearchLogEntries :many
SELECT agent_id, source, level, message, count, first_seen, last_seen, received_at
FROM log_entries
WHERE agent_id = $1
  AND received_at >= $2
  AND received_at < $3
  AND search_vec @@ plainto_tsquery('english', $4)
ORDER BY received_at DESC
LIMIT $5;
