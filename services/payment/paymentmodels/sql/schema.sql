-- ======
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE tier_id AS ENUM ('t0', 't1', 't2');


CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    email TEXT NOT NULL UNIQUE,
    tier_id tier_id NOT NULL DEFAULT 't0',
    stripe_customer_id TEXT UNIQUE,
    stripe_subscription_id TEXT UNIQUE
);


-- ======
CREATE TABLE otps (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL,
    user_id uuid NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
);
