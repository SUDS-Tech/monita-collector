-- name: CreateAlertRule :one
INSERT INTO alert_rules (org_id, name, target, condition, severity, notification_channels)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, org_id, name, target, condition, severity, notification_channels, enabled, created_at, updated_at;

-- name: GetAlertRule :one
SELECT id, org_id, name, target, condition, severity, notification_channels, enabled, created_at, updated_at
FROM alert_rules WHERE id = $1 AND org_id = $2 LIMIT 1;

-- name: ListAlertRules :many
SELECT id, name, target, severity, enabled, created_at, updated_at
FROM alert_rules WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdateAlertRule :one
UPDATE alert_rules
SET name = $3, condition = $4, severity = $5, notification_channels = $6, enabled = $7, updated_at = now()
WHERE id = $1 AND org_id = $2
RETURNING id, org_id, name, target, condition, severity, notification_channels, enabled, created_at, updated_at;

-- name: DeleteAlertRule :exec
DELETE FROM alert_rules WHERE id = $1 AND org_id = $2;

-- name: ListAlertEvents :many
SELECT id, rule_id, agent_id, fired_at, resolved_at, value_at_fire, status
FROM alert_events WHERE rule_id = $1 ORDER BY fired_at DESC LIMIT $2;

-- name: ListFiringEvents :many
SELECT ae.id, ae.rule_id, ae.agent_id, ae.fired_at, ae.status
FROM alert_events ae
JOIN alert_rules ar ON ar.id = ae.rule_id
WHERE ar.org_id = $1 AND ae.status = 'firing'
ORDER BY ae.fired_at DESC
LIMIT $2;

-- name: ResolveAlertEvent :exec
UPDATE alert_events SET resolved_at = now(), status = 'resolved' WHERE id = $1;
