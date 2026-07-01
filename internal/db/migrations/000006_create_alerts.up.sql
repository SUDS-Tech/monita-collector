CREATE TYPE alert_target AS ENUM ('metric', 'log_pattern', 'heartbeat', 'security');
CREATE TYPE alert_severity AS ENUM ('info', 'warning', 'critical');
CREATE TYPE alert_event_status AS ENUM ('firing', 'resolved');

CREATE TABLE alert_rules (
    id                    UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                UUID           NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                  TEXT           NOT NULL,
    target                alert_target   NOT NULL,
    condition             JSONB          NOT NULL,
    severity              alert_severity NOT NULL DEFAULT 'warning',
    notification_channels JSONB          NOT NULL DEFAULT '[]',
    escalate_after        INTERVAL,
    escalate_to           JSONB,
    enabled               BOOLEAN        NOT NULL DEFAULT true,
    created_at            TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE TABLE alert_events (
    id            UUID               PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id       UUID               NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    agent_id      UUID               NOT NULL REFERENCES agents(id)      ON DELETE CASCADE,
    fired_at      TIMESTAMPTZ        NOT NULL DEFAULT now(),
    resolved_at   TIMESTAMPTZ,
    value_at_fire JSONB,
    status        alert_event_status NOT NULL DEFAULT 'firing'
);

CREATE INDEX idx_alert_rules_org_id    ON alert_rules(org_id);
CREATE INDEX idx_alert_events_rule_id  ON alert_events(rule_id);
CREATE INDEX idx_alert_events_agent_id ON alert_events(agent_id, fired_at DESC);
CREATE INDEX idx_alert_events_firing   ON alert_events(status) WHERE status = 'firing';
