-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS store_statuses (
    id BIGSERIAL PRIMARY KEY,
    business_date DATE NOT NULL,
    store_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    is_open BOOLEAN NOT NULL DEFAULT FALSE,
    cash_open NUMERIC(15, 2) NULL,
    cash_close NUMERIC(15, 2) NULL,
    expected_cash_at_close NUMERIC(15, 2) NULL,
    cash_difference NUMERIC(15, 2) NULL,
    scheduled_close_time TIME NULL,
    opened_by BIGINT NULL,
    closed_by BIGINT NULL,
    opened_at TIMESTAMPTZ NULL,
    closed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_store_statuses_business_date_branch
        UNIQUE (business_date, branch_id),

    CONSTRAINT chk_store_statuses_cash_open
        CHECK (cash_open IS NULL OR cash_open >= 0),

    CONSTRAINT chk_store_statuses_cash_close
        CHECK (cash_close IS NULL OR cash_close >= 0),

    CONSTRAINT fk_store_statuses_store_id
        FOREIGN KEY (store_id)
        REFERENCES stores(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_store_statuses_branch_id
        FOREIGN KEY (branch_id)
        REFERENCES branches(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_store_statuses_opened_by
        FOREIGN KEY (opened_by)
        REFERENCES users(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_store_statuses_closed_by
        FOREIGN KEY (closed_by)
        REFERENCES users(id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_store_statuses_store_id ON store_statuses(store_id);
CREATE INDEX IF NOT EXISTS idx_store_statuses_branch_id ON store_statuses(branch_id);
CREATE INDEX IF NOT EXISTS idx_store_statuses_business_date ON store_statuses(business_date);
CREATE INDEX IF NOT EXISTS idx_store_statuses_is_open ON store_statuses(is_open);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_store_statuses_is_open;
DROP INDEX IF EXISTS idx_store_statuses_business_date;
DROP INDEX IF EXISTS idx_store_statuses_branch_id;
DROP INDEX IF EXISTS idx_store_statuses_store_id;

DROP TABLE IF EXISTS store_statuses;

-- +goose StatementEnd
