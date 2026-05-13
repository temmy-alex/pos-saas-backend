package repositories

import (
	"context"
	"database/sql"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/models"
	"pos-saas-backend/internal/requests"
)

type CategoryRepository struct {
	DB *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{
		DB: db,
	}
}

func (r *CategoryRepository) FindAll(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT
			c.id,
			c.store_id,
			s.name AS store_name,
			c.name,
			c.code,
			c.description,
			c.is_active,
			c.created_at,
			c.updated_at
		FROM categories c
		INNER JOIN stores s ON s.id = c.store_id
		WHERE c.deleted_at IS NULL
		ORDER BY c.id ASC
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]models.Category, 0)

	for rows.Next() {
		var category models.Category
		var description sql.NullString

		err := rows.Scan(
			&category.ID,
			&category.StoreID,
			&category.StoreName,
			&category.Name,
			&category.Code,
			&description,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		category.Description = helpers.NullableString(description)

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, id int64) (*models.Category, error) {
	query := `
		SELECT
			c.id,
			c.store_id,
			s.name AS store_name,
			c.name,
			c.code,
			c.description,
			c.is_active,
			c.created_at,
			c.updated_at
		FROM categories c
		INNER JOIN stores s ON s.id = c.store_id
		WHERE c.id = $1
		AND c.deleted_at IS NULL
		LIMIT 1
	`

	var category models.Category
	var description sql.NullString

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.StoreID,
		&category.StoreName,
		&category.Name,
		&category.Code,
		&description,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	category.Description = helpers.NullableString(description)

	return &category, nil
}

func (r *CategoryRepository) Create(ctx context.Context, request requests.CategoryRequest) (*models.Category, error) {
	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	query := `
		INSERT INTO categories (
			store_id,
			name,
			code,
			description,
			is_active
		) VALUES (
			$1,
			$2,
			$3,
			NULLIF($4, ''),
			$5
		)
		RETURNING id
	`

	var categoryID int64

	err := r.DB.QueryRowContext(
		ctx,
		query,
		request.StoreID,
		request.Name,
		request.Code,
		request.Description,
		isActive,
	).Scan(&categoryID)

	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, categoryID)
}

func (r *CategoryRepository) Update(ctx context.Context, id int64, request requests.CategoryRequest) (*models.Category, error) {
	currentCategory, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if currentCategory == nil {
		return nil, nil
	}

	isActive := currentCategory.IsActive
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	query := `
		UPDATE categories
		SET
			store_id = $1,
			name = $2,
			code = $3,
			description = NULLIF($4, ''),
			is_active = $5,
			updated_at = NOW()
		WHERE id = $6
		AND deleted_at IS NULL
		RETURNING id
	`

	var categoryID int64

	err = r.DB.QueryRowContext(
		ctx,
		query,
		request.StoreID,
		request.Name,
		request.Code,
		request.Description,
		isActive,
		id,
	).Scan(&categoryID)

	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, categoryID)
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) (bool, error) {
	query := `
		UPDATE categories
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
