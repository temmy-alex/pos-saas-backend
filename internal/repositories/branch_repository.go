package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/models"
	"pos-saas-backend/internal/requests"
)

type BranchRepository struct {
	DB *sql.DB
}

func NewBranchRepository(db *sql.DB) *BranchRepository {
	return &BranchRepository{
		DB: db,
	}
}

func (r *BranchRepository) FindAll(ctx context.Context, storeID int64, branchID int64) ([]models.Branch, error) {
	whereClauses := []string{
		"b.deleted_at IS NULL",
	}

	args := make([]interface{}, 0)
	argPosition := 1

	if storeID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("b.store_id = $%d", argPosition))
		args = append(args, storeID)
		argPosition++
	}

	if branchID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("b.id = $%d", argPosition))
		args = append(args, branchID)
		argPosition++
	}

	query := fmt.Sprintf(`
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
		WHERE %s
		ORDER BY b.id ASC
	`, strings.Join(whereClauses, " AND "))

	rows, err := r.DB.QueryContext(ctx, query, args...)
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

func (r *BranchRepository) FindByID(ctx context.Context, id int64) (*models.Branch, error) {
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
		WHERE b.id = $1
		AND b.deleted_at IS NULL
		LIMIT 1
	`

	var branch models.Branch
	var address sql.NullString
	var phone sql.NullString

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
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
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	branch.Address = helpers.NullableString(address)
	branch.Phone = helpers.NullableString(phone)

	return &branch, nil
}

func (r *BranchRepository) Create(ctx context.Context, request requests.BranchRequest) (*models.Branch, error) {
	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	query := `
		INSERT INTO branches (
			store_id,
			name,
			code,
			address,
			phone,
			is_active
		) VALUES (
			$1,
			$2,
			$3,
			NULLIF($4, ''),
			NULLIF($5, ''),
			$6
		)
		RETURNING
			id
	`

	var branchID int64

	err := r.DB.QueryRowContext(
		ctx,
		query,
		request.StoreID,
		request.Name,
		request.Code,
		request.Address,
		request.Phone,
		isActive,
	).Scan(&branchID)

	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, branchID)
}

func (r *BranchRepository) Update(ctx context.Context, id int64, request requests.BranchRequest) (*models.Branch, error) {
	currentBranch, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if currentBranch == nil {
		return nil, nil
	}

	isActive := currentBranch.IsActive
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	query := `
		UPDATE branches
		SET
			store_id = $1,
			name = $2,
			code = $3,
			address = NULLIF($4, ''),
			phone = NULLIF($5, ''),
			is_active = $6,
			updated_at = NOW()
		WHERE id = $7
		AND deleted_at IS NULL
		RETURNING
			id
	`

	var branchID int64

	err = r.DB.QueryRowContext(
		ctx,
		query,
		request.StoreID,
		request.Name,
		request.Code,
		request.Address,
		request.Phone,
		isActive,
		id,
	).Scan(&branchID)

	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, branchID)
}

func (r *BranchRepository) Delete(ctx context.Context, id int64) (bool, error) {
	query := `
		UPDATE branches
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
