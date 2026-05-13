package repositories

import (
	"context"
	"database/sql"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/models"
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
