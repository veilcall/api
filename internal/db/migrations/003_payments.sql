CREATE TABLE IF NOT EXISTS payments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    monero_address VARCHAR(106) NOT NULL,
    monero_amount  NUMERIC(20,12) NOT NULL,
    plan           VARCHAR(4) NOT NULL,
    country        CHAR(2) NOT NULL,
    status         VARCHAR(12) NOT NULL DEFAULT 'pending',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at   TIMESTAMPTZ,
    number_id      UUID REFERENCES phone_numbers(id)
);

CREATE INDEX IF NOT EXISTS idx_payments_pending
    ON payments(status)
    WHERE status = 'pending';
