-- +goose Up
-- +goose StatementBegin

INSERT INTO stores (
    name,
    code,
    address,
    phone,
    is_active
) VALUES (
    'Second Store',
    'SECOND',
    'Second store for role testing',
    '081111111111',
    TRUE
)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    address = EXCLUDED.address,
    phone = EXCLUDED.phone,
    is_active = TRUE,
    updated_at = NOW();

INSERT INTO branches (
    store_id,
    name,
    code,
    address,
    phone,
    is_active
) VALUES (
    (SELECT id FROM stores WHERE code = 'SECOND' LIMIT 1),
    'Second Branch',
    'SECOND-BRANCH',
    'Second branch for role testing',
    '082222222222',
    TRUE
)
ON CONFLICT (code) DO UPDATE SET
    store_id = EXCLUDED.store_id,
    name = EXCLUDED.name,
    address = EXCLUDED.address,
    phone = EXCLUDED.phone,
    is_active = TRUE,
    updated_at = NOW();

INSERT INTO users (
    store_id,
    branch_id,
    name,
    email,
    password_hash,
    role,
    is_active
) VALUES (
    (SELECT id FROM stores WHERE code = 'MAIN' LIMIT 1),
    NULL,
    'Admin Main Store',
    'admin.store@possaas.test',
    (SELECT password_hash FROM users WHERE email = 'admin@possaas.test' LIMIT 1),
    'admin',
    TRUE
)
ON CONFLICT (email) DO UPDATE SET
    store_id = EXCLUDED.store_id,
    branch_id = EXCLUDED.branch_id,
    name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    is_active = TRUE,
    updated_at = NOW();

INSERT INTO users (
    store_id,
    branch_id,
    name,
    email,
    password_hash,
    role,
    is_active
) VALUES (
    (SELECT id FROM stores WHERE code = 'MAIN' LIMIT 1),
    (SELECT id FROM branches WHERE code = 'MAIN-BRANCH' LIMIT 1),
    'Cashier Main Branch',
    'cashier.branch@possaas.test',
    (SELECT password_hash FROM users WHERE email = 'admin@possaas.test' LIMIT 1),
    'cashier',
    TRUE
)
ON CONFLICT (email) DO UPDATE SET
    store_id = EXCLUDED.store_id,
    branch_id = EXCLUDED.branch_id,
    name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    is_active = TRUE,
    updated_at = NOW();

INSERT INTO users (
    store_id,
    branch_id,
    name,
    email,
    password_hash,
    role,
    is_active
) VALUES (
    (SELECT id FROM stores WHERE code = 'SECOND' LIMIT 1),
    NULL,
    'Admin Second Store',
    'admin.second@possaas.test',
    (SELECT password_hash FROM users WHERE email = 'admin@possaas.test' LIMIT 1),
    'admin',
    TRUE
)
ON CONFLICT (email) DO UPDATE SET
    store_id = EXCLUDED.store_id,
    branch_id = EXCLUDED.branch_id,
    name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    is_active = TRUE,
    updated_at = NOW();

INSERT INTO users (
    store_id,
    branch_id,
    name,
    email,
    password_hash,
    role,
    is_active
) VALUES (
    (SELECT id FROM stores WHERE code = 'SECOND' LIMIT 1),
    (SELECT id FROM branches WHERE code = 'SECOND-BRANCH' LIMIT 1),
    'Cashier Second Branch',
    'cashier.second@possaas.test',
    (SELECT password_hash FROM users WHERE email = 'admin@possaas.test' LIMIT 1),
    'cashier',
    TRUE
)
ON CONFLICT (email) DO UPDATE SET
    store_id = EXCLUDED.store_id,
    branch_id = EXCLUDED.branch_id,
    name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    is_active = TRUE,
    updated_at = NOW();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM users
WHERE email IN (
    'admin.store@possaas.test',
    'cashier.branch@possaas.test',
    'admin.second@possaas.test',
    'cashier.second@possaas.test'
);

DELETE FROM branches
WHERE code = 'SECOND-BRANCH';

DELETE FROM stores
WHERE code = 'SECOND';

-- +goose StatementEnd