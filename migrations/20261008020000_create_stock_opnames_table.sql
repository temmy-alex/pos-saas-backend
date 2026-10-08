-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS stock_opnames (
    id BIGSERIAL PRIMARY KEY,
    opname_number VARCHAR(100) NOT NULL UNIQUE,
    store_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    counted_by BIGINT NOT NULL,
    completed_by BIGINT NULL,
    cancelled_by BIGINT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'draft',
    note VARCHAR(255) NULL,
    cancel_reason VARCHAR(255) NULL,
    completed_at TIMESTAMPTZ NULL,
    cancelled_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_stock_opnames_status
        CHECK (status IN ('draft', 'completed', 'cancelled')),

    CONSTRAINT fk_stock_opnames_store_id
        FOREIGN KEY (store_id)
        REFERENCES stores(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_stock_opnames_branch_id
        FOREIGN KEY (branch_id)
        REFERENCES branches(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_stock_opnames_counted_by
        FOREIGN KEY (counted_by)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_stock_opnames_completed_by
        FOREIGN KEY (completed_by)
        REFERENCES users(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_stock_opnames_cancelled_by
        FOREIGN KEY (cancelled_by)
        REFERENCES users(id)
        ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS stock_opname_details (
    id BIGSERIAL PRIMARY KEY,
    stock_opname_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    store_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    old_stock INTEGER NOT NULL,
    new_stock INTEGER NOT NULL,
    difference INTEGER NOT NULL,
    note VARCHAR(255) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_stock_opname_details_product
        UNIQUE (stock_opname_id, product_id),

    CONSTRAINT chk_stock_opname_details_old_stock
        CHECK (old_stock >= 0),

    CONSTRAINT chk_stock_opname_details_new_stock
        CHECK (new_stock >= 0),

    CONSTRAINT fk_stock_opname_details_opname_id
        FOREIGN KEY (stock_opname_id)
        REFERENCES stock_opnames(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_stock_opname_details_product_id
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_stock_opname_details_store_id
        FOREIGN KEY (store_id)
        REFERENCES stores(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_stock_opname_details_branch_id
        FOREIGN KEY (branch_id)
        REFERENCES branches(id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_stock_opnames_store_id ON stock_opnames(store_id);
CREATE INDEX IF NOT EXISTS idx_stock_opnames_branch_id ON stock_opnames(branch_id);
CREATE INDEX IF NOT EXISTS idx_stock_opnames_counted_by ON stock_opnames(counted_by);
CREATE INDEX IF NOT EXISTS idx_stock_opnames_status ON stock_opnames(status);
CREATE INDEX IF NOT EXISTS idx_stock_opnames_created_at ON stock_opnames(created_at);

CREATE INDEX IF NOT EXISTS idx_stock_opname_details_opname_id ON stock_opname_details(stock_opname_id);
CREATE INDEX IF NOT EXISTS idx_stock_opname_details_product_id ON stock_opname_details(product_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_stock_opname_details_product_id;
DROP INDEX IF EXISTS idx_stock_opname_details_opname_id;

DROP INDEX IF EXISTS idx_stock_opnames_created_at;
DROP INDEX IF EXISTS idx_stock_opnames_status;
DROP INDEX IF EXISTS idx_stock_opnames_counted_by;
DROP INDEX IF EXISTS idx_stock_opnames_branch_id;
DROP INDEX IF EXISTS idx_stock_opnames_store_id;

DROP TABLE IF EXISTS stock_opname_details;
DROP TABLE IF EXISTS stock_opnames;

-- +goose StatementEnd
