-- name: CreateUser :one
INSERT INTO users (email, name, password_hash)
VALUES ($1, $2, $3)
RETURNING id, email, name, created_at;

-- name: GetUserForLogin :one
-- Returns password_hash so the service can verify it. Not used anywhere else.
SELECT id, email, name, password_hash
FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT id, email, name, created_at
FROM users
WHERE id = $1
LIMIT 1;

-- name: CreateOrganization :one
INSERT INTO organizations (name, slug, owner_id)
VALUES ($1, $2, $3)
RETURNING id, name, slug, owner_id, subscription_status, trial_ends_at, created_at;

-- name: GetOrganizationByOwnerID :one
SELECT id, name, slug, owner_id, subscription_status, stripe_customer_id, stripe_subscription_id
FROM organizations
WHERE owner_id = $1
LIMIT 1;
