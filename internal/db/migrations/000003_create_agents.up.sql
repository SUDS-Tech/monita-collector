CREATE TABLE agents (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id            UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name              TEXT        NOT NULL,
    hostname          TEXT        NOT NULL,
    tags              JSONB       NOT NULL DEFAULT '{}',
    token_hash        TEXT        NOT NULL UNIQUE,
    signing_key_hash  TEXT        NOT NULL,
    fingerprint_hash  TEXT,
    fingerprint_drift BOOLEAN     NOT NULL DEFAULT false,
    frozen            BOOLEAN     NOT NULL DEFAULT false,
    rotation_required BOOLEAN     NOT NULL DEFAULT false,
    expires_at        TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '1 year',
    revoked           BOOLEAN     NOT NULL DEFAULT false,
    last_seen_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_agents_org_id     ON agents(org_id);
CREATE INDEX idx_agents_token_hash ON agents(token_hash);
