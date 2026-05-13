-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS branches (
    id BIGSERIAL PRIMARY KEY,
    store_id BIGINT NOT NULL,
    name VARCHAR(150) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    address TEXT NULL,
    phone VARCHAR(50) NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT fk_branches_store_id
        FOREIGN KEY (store_id)
        REFERENCES stores(id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_branches_store_id ON branches(store_id);
CREATE INDEX IF NOT EXISTS idx_branches_code ON branches(code);
CREATE INDEX IF NOT EXISTS idx_branches_is_active ON branches(is_active);
CREATE INDEX IF NOT EXISTS idx_branches_deleted_at ON branches(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_branches_deleted_at;
DROP INDEX IF EXISTS idx_branches_is_active;
DROP INDEX IF EXISTS idx_branches_code;
DROP INDEX IF EXISTS idx_branches_store_id;

DROP TABLE IF EXISTS branches;

-- +goose StatementEnd