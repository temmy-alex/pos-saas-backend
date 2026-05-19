-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,

    store_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    cashier_id BIGINT NOT NULL,

    transaction_number VARCHAR(100) NOT NULL UNIQUE,
    customer_name VARCHAR(150) NULL,

    payment_method VARCHAR(50) NOT NULL DEFAULT 'cash',

    subtotal NUMERIC(15, 2) NOT NULL DEFAULT 0,
    discount_total NUMERIC(15, 2) NOT NULL DEFAULT 0,
    grand_total NUMERIC(15, 2) NOT NULL DEFAULT 0,

    cash_amount NUMERIC(15, 2) NOT NULL DEFAULT 0,
    transfer_amount NUMERIC(15, 2) NOT NULL DEFAULT 0,
    change_amount NUMERIC(15, 2) NOT NULL DEFAULT 0,

    notes TEXT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'paid',

    transaction_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT fk_transactions_store_id
        FOREIGN KEY (store_id)
        REFERENCES stores(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_transactions_branch_id
        FOREIGN KEY (branch_id)
        REFERENCES branches(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_transactions_cashier_id
        FOREIGN KEY (cashier_id)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_transactions_payment_method
        CHECK (payment_method IN ('cash', 'transfer', 'mixed')),

    CONSTRAINT chk_transactions_status
        CHECK (status IN ('paid', 'void'))
);

CREATE TABLE IF NOT EXISTS transaction_items (
    id BIGSERIAL PRIMARY KEY,

    transaction_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,

    product_name_snapshot VARCHAR(150) NOT NULL,
    product_sku_snapshot VARCHAR(100) NOT NULL,

    qty INTEGER NOT NULL DEFAULT 1,
    price NUMERIC(15, 2) NOT NULL DEFAULT 0,
    discount NUMERIC(15, 2) NOT NULL DEFAULT 0,
    subtotal NUMERIC(15, 2) NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT fk_transaction_items_transaction_id
        FOREIGN KEY (transaction_id)
        REFERENCES transactions(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_transaction_items_product_id
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_transactions_store_id ON transactions(store_id);
CREATE INDEX IF NOT EXISTS idx_transactions_branch_id ON transactions(branch_id);
CREATE INDEX IF NOT EXISTS idx_transactions_cashier_id ON transactions(cashier_id);
CREATE INDEX IF NOT EXISTS idx_transactions_transaction_number ON transactions(transaction_number);
CREATE INDEX IF NOT EXISTS idx_transactions_transaction_date ON transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_deleted_at ON transactions(deleted_at);

CREATE INDEX IF NOT EXISTS idx_transaction_items_transaction_id ON transaction_items(transaction_id);
CREATE INDEX IF NOT EXISTS idx_transaction_items_product_id ON transaction_items(product_id);
CREATE INDEX IF NOT EXISTS idx_transaction_items_deleted_at ON transaction_items(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_transaction_items_deleted_at;
DROP INDEX IF EXISTS idx_transaction_items_product_id;
DROP INDEX IF EXISTS idx_transaction_items_transaction_id;

DROP INDEX IF EXISTS idx_transactions_deleted_at;
DROP INDEX IF EXISTS idx_transactions_status;
DROP INDEX IF EXISTS idx_transactions_transaction_date;
DROP INDEX IF EXISTS idx_transactions_transaction_number;
DROP INDEX IF EXISTS idx_transactions_cashier_id;
DROP INDEX IF EXISTS idx_transactions_branch_id;
DROP INDEX IF EXISTS idx_transactions_store_id;

DROP TABLE IF EXISTS transaction_items;
DROP TABLE IF EXISTS transactions;

-- +goose StatementEnd