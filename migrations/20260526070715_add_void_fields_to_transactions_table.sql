-- +goose Up
-- +goose StatementBegin

ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS void_reason TEXT NULL;

ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS voided_at TIMESTAMPTZ NULL;

ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS voided_by BIGINT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_transactions_voided_by'
    ) THEN
        ALTER TABLE transactions
        ADD CONSTRAINT fk_transactions_voided_by
            FOREIGN KEY (voided_by)
            REFERENCES users(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_transactions_voided_by ON transactions(voided_by);
CREATE INDEX IF NOT EXISTS idx_transactions_voided_at ON transactions(voided_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_transactions_voided_at;
DROP INDEX IF EXISTS idx_transactions_voided_by;

ALTER TABLE transactions
DROP CONSTRAINT IF EXISTS fk_transactions_voided_by;

ALTER TABLE transactions
DROP COLUMN IF EXISTS voided_by;

ALTER TABLE transactions
DROP COLUMN IF EXISTS voided_at;

ALTER TABLE transactions
DROP COLUMN IF EXISTS void_reason;

-- +goose StatementEnd