CREATE TABLE IF NOT EXISTS phone_numbers (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    telnyx_number VARCHAR(20) NOT NULL,
    country       CHAR(2) NOT NULL,
    plan          VARCHAR(4) NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL,
    released      BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_numbers_expires
    ON phone_numbers(expires_at)
    WHERE released = FALSE;
