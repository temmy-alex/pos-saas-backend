-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS customers (
    id BIGSERIAL PRIMARY KEY,
    store_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    name VARCHAR(150) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    email VARCHAR(150) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT fk_customers_store_id
        FOREIGN KEY (store_id)
        REFERENCES stores(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_customers_branch_id
        FOREIGN KEY (branch_id)
        REFERENCES branches(id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_customers_store_id ON customers(store_id);
CREATE INDEX IF NOT EXISTS idx_customers_branch_id ON customers(branch_id);
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
CREATE INDEX IF NOT EXISTS idx_customers_deleted_at ON customers(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_customers_deleted_at;
DROP INDEX IF EXISTS idx_customers_email;
DROP INDEX IF EXISTS idx_customers_phone;
DROP INDEX IF EXISTS idx_customers_branch_id;
DROP INDEX IF EXISTS idx_customers_store_id;

DROP TABLE IF EXISTS customers;

-- +goose StatementEnd
