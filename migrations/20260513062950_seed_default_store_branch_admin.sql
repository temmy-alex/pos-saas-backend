-- +goose Up
-- +goose StatementBegin

INSERT INTO stores (
    name,
    code,
    address,
    phone,
    is_active
) VALUES (
    'Main Store',
    'MAIN',
    'Default main store',
    '081234567890',
    TRUE
)
ON CONFLICT (code) DO NOTHING;

INSERT INTO branches (
    store_id,
    name,
    code,
    address,
    phone,
    is_active
) VALUES (
    (SELECT id FROM stores WHERE code = 'MAIN' LIMIT 1),
    'Main Branch',
    'MAIN-BRANCH',
    'Default main branch',
    '081234567890',
    TRUE
)
ON CONFLICT (code) DO NOTHING;

INSERT INTO users (
    store_id,
    branch_id,
    name,
    email,
    password_hash,
    role,
    is_active
) VALUES (
    NULL,
    NULL,
    'Super Admin',
    'admin@possaas.test',
    '$2a$10$BzUIGTKiNclwSPpPcVExSu5/rtOCIpKvrrhelA3zZBfQA7ItZWgLy',
    'superadmin',
    TRUE
)
ON CONFLICT (email) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM users
WHERE email = 'admin@possaas.test';

DELETE FROM branches
WHERE code = 'MAIN-BRANCH';

DELETE FROM stores
WHERE code = 'MAIN';

-- +goose StatementEnd