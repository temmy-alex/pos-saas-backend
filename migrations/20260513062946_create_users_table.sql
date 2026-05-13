-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    store_id BIGINT NULL,
    branch_id BIGINT NULL,
    name VARCHAR(150) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'cashier',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT chk_users_role
        CHECK (role IN ('superadmin', 'admin', 'cashier')),

    CONSTRAINT fk_users_store_id
        FOREIGN KEY (store_id)
        REFERENCES stores(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_users_branch_id
        FOREIGN KEY (branch_id)
        REFERENCES branches(id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_users_store_id ON users(store_id);
CREATE INDEX IF NOT EXISTS idx_users_branch_id ON users(branch_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_branch_id;
DROP INDEX IF EXISTS idx_users_store_id;

DROP TABLE IF EXISTS users;

-- +goose StatementEnd