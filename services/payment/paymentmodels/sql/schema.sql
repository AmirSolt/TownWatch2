-- ======
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    email TEXT NOT NULL UNIQUE,
    -- tier INT NOT NULL DEFAULT 0,
    stripe_customer_id TEXT UNIQUE,
    -- stripe_subscription_id TEXT UNIQUE,
    user_id uuid NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE FUNCTION customer_insert() RETURNS trigger AS $$
    BEGIN
        INSERT INTO customers(email, user_id)
		 VALUES(NEW.email, NEW.id);
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER create_customer_on_user AFTER INSERT OR UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION customer_insert();
