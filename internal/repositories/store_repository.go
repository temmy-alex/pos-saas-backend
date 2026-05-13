package repositories

import (
	"context"
	"database/sql"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/models"
	"pos-saas-backend/internal/requests"
)

type StoreRepository struct {
	DB *sql.DB
}

func NewStoreRepository(db *sql.DB) *StoreRepository {
	return &StoreRepository{
		DB: db,
	}
}

func (r *StoreRepository) FindAll(ctx context.Context) ([]models.Store, error) {
	query := `
		SELECT
			id,
			name,
			code,
			address,
			phone,
			is_active,
			created_at,
			updated_at
		FROM stores
		WHERE deleted_at IS NULL
		ORDER BY id ASC
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stores := make([]models.Store, 0)

	for rows.Next() {
		var store models.Store
		var address sql.NullString
		var phone sql.NullString

		err := rows.Scan(
			&store.ID,
			&store.Name,
			&store.Code,
			&address,
			&phone,
			&store.IsActive,
			&store.CreatedAt,
			&store.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		store.Address = helpers.NullableString(address)
		store.Phone = helpers.NullableString(phone)

		stores = append(stores, store)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stores, nil
}

func (r *StoreRepository) FindByID(ctx context.Context, id int64) (*models.Store, error) {
	query := `
		SELECT
			id,
			name,
			code,
			address,
			phone,
			is_active,
			created_at,
			updated_at
		FROM stores
		WHERE id = $1
		AND deleted_at IS NULL
		LIMIT 1
	`

	var store models.Store
	var address sql.NullString
	var phone sql.NullString

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&store.ID,
		&store.Name,
		&store.Code,
		&address,
		&phone,
		&store.IsActive,
		&store.CreatedAt,
		&store.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	store.Address = helpers.NullableString(address)
	store.Phone = helpers.NullableString(phone)

	return &store, nil
}

func (r *StoreRepository) Create(ctx context.Context, request requests.StoreRequest) (*models.Store, error) {
	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	query := `
		INSERT INTO stores (
			name,
			code,
			address,
			phone,
			is_active
		) VALUES (
			$1,
			$2,
			NULLIF($3, ''),
			NULLIF($4, ''),
			$5
		)
		RETURNING
			id,
			name,
			code,
			address,
			phone,
			is_active,
			created_at,
			updated_at
	`

	var store models.Store
	var address sql.NullString
	var phone sql.NullString

	err := r.DB.QueryRowContext(
		ctx,
		query,
		request.Name,
		request.Code,
		request.Address,
		request.Phone,
		isActive,
	).Scan(
		&store.ID,
		&store.Name,
		&store.Code,
		&address,
		&phone,
		&store.IsActive,
		&store.CreatedAt,
		&store.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	store.Address = helpers.NullableString(address)
	store.Phone = helpers.NullableString(phone)

	return &store, nil
}

func (r *StoreRepository) Update(ctx context.Context, id int64, request requests.StoreRequest) (*models.Store, error) {
	currentStore, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if currentStore == nil {
		return nil, nil
	}

	isActive := currentStore.IsActive
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	query := `
		UPDATE stores
		SET
			name = $1,
			code = $2,
			address = NULLIF($3, ''),
			phone = NULLIF($4, ''),
			is_active = $5,
			updated_at = NOW()
		WHERE id = $6
		AND deleted_at IS NULL
		RETURNING
			id,
			name,
			code,
			address,
			phone,
			is_active,
			created_at,
			updated_at
	`

	var store models.Store
	var address sql.NullString
	var phone sql.NullString

	err = r.DB.QueryRowContext(
		ctx,
		query,
		request.Name,
		request.Code,
		request.Address,
		request.Phone,
		isActive,
		id,
	).Scan(
		&store.ID,
		&store.Name,
		&store.Code,
		&address,
		&phone,
		&store.IsActive,
		&store.CreatedAt,
		&store.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	store.Address = helpers.NullableString(address)
	store.Phone = helpers.NullableString(phone)

	return &store, nil
}

func (r *StoreRepository) Delete(ctx context.Context, id int64) (bool, error) {
	query := `
		UPDATE stores
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		AND deleted_at IS NULL
	`

	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}
