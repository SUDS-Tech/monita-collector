-- name: CreateAgent :one
INSERT INTO agents (org_id, name, hostname, tags, token_hash, signing_key_hash)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, org_id, name, hostname, tags, frozen, revoked, rotation_required, expires_at, last_seen_at, created_at;

-- name: GetAgentByID :one
SELECT id, org_id, name, hostname, tags, frozen, revoked, rotation_required, expires_at, last_seen_at, created_at
FROM agents WHERE id = $1 AND org_id = $2 LIMIT 1;

-- name: ListAgentsByOrgID :many
SELECT id, name, hostname, tags, frozen, revoked, rotation_required, expires_at, last_seen_at, created_at
FROM agents WHERE org_id = $1 ORDER BY created_at DESC;

-- name: GetAgentByTokenHash :one
SELECT id, org_id, signing_key_hash, fingerprint_hash, fingerprint_drift, frozen, revoked, expires_at
FROM agents WHERE token_hash = $1 LIMIT 1;

-- name: UpdateAgentLastSeen :exec
UPDATE agents SET last_seen_at = now() WHERE id = $1;

-- name: FreezeAgent :exec
UPDATE agents SET frozen = true WHERE id = $1;

-- name: SetFingerprintHash :exec
UPDATE agents SET fingerprint_hash = $2 WHERE id = $1;

-- name: SetFingerprintDrift :exec
UPDATE agents SET fingerprint_drift = true WHERE id = $1;

-- name: RevokeAgent :exec
UPDATE agents SET revoked = true WHERE id = $1 AND org_id = $2;

-- name: DeleteAgent :exec
DELETE FROM agents WHERE id = $1 AND org_id = $2;