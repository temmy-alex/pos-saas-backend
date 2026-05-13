package repositories

import (
	"context"
	"database/sql"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/models"
)

type BranchRepository struct {
	DB *sql.DB
}

func NewBranchRepository(db *sql.DB) *BranchRepository {
	return &BranchRepository{
		DB: db,
	}
}

func (r *BranchRepository) FindAll(ctx context.Context) ([]models.Branch, error) {
	query := `
		SELECT
			b.id,
			b.store_id,
			s.name AS store_name,
			b.name,
			b.code,
			b.address,
			b.phone,
			b.is_active,
			b.created_at,
			b.updated_at
		FROM branches b
		INNER JOIN stores s ON s.id = b.store_id
		WHERE b.deleted_at IS NULL
		ORDER BY b.id ASC
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	branches := make([]models.Branch, 0)

	for rows.Next() {
		var branch models.Branch
		var address sql.NullString
		var phone sql.NullString

		err := rows.Scan(
			&branch.ID,
			&branch.StoreID,
			&branch.StoreName,
			&branch.Name,
			&branch.Code,
			&address,
			&phone,
			&branch.IsActive,
			&branch.CreatedAt,
			&branch.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		branch.Address = helpers.NullableString(address)
		branch.Phone = helpers.NullableString(phone)

		branches = append(branches, branch)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return branches, nil
}
