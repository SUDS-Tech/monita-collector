-- FK from partitioned table to agents requires PostgreSQL 15+.
CREATE TABLE log_entries (
    agent_id    UUID        NOT NULL,
    source      TEXT        NOT NULL,
    level       TEXT        NOT NULL,
    message     TEXT        NOT NULL,
    count       INTEGER     NOT NULL DEFAULT 1,
    first_seen  TIMESTAMPTZ NOT NULL,
    last_seen   TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    search_vec  TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', message)) STORED
) PARTITION BY RANGE (received_at);

CREATE TABLE log_entries_default PARTITION OF log_entries DEFAULT;

CREATE INDEX idx_log_entries_brin     ON log_entries USING BRIN (received_at);
CREATE INDEX idx_log_entries_agent_id ON log_entries (agent_id, received_at DESC);
CREATE INDEX idx_log_entries_search   ON log_entries USING GIN (search_vec);
