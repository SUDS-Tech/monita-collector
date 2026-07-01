CREATE TYPE subscription_status AS ENUM (
    'trialing',
    'active',
    'past_due',
    'unpaid',
    'canceled'
);

CREATE TABLE organizations (
    id                     UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    name                   TEXT                NOT NULL,
    slug                   TEXT                NOT NULL UNIQUE,
    owner_id               UUID                NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stripe_customer_id     TEXT                UNIQUE,
    stripe_subscription_id TEXT,
    subscription_status    subscription_status NOT NULL DEFAULT 'trialing',
    trial_ends_at          TIMESTAMPTZ         NOT NULL DEFAULT now() + INTERVAL '14 days',
    created_at             TIMESTAMPTZ         NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ         NOT NULL DEFAULT now()
);

CREATE INDEX idx_organizations_owner_id           ON organizations(owner_id);
CREATE INDEX idx_organizations_stripe_customer_id ON organizations(stripe_customer_id)
    WHERE stripe_customer_id IS NOT NULL;
