package repositories

import (
	"context"
	"database/sql"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT
			id,
			store_id,
			branch_id,
			name,
			email,
			password_hash,
			role,
			is_active,
			last_login_at,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
		AND deleted_at IS NULL
		LIMIT 1
	`

	var user models.User
	var storeID sql.NullInt64
	var branchID sql.NullInt64
	var lastLoginAt sql.NullTime

	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&storeID,
		&branchID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	user.StoreID = helpers.NullableInt64(storeID)
	user.BranchID = helpers.NullableInt64(branchID)
	user.LastLoginAt = helpers.NullableTime(lastLoginAt)

	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET
			last_login_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		AND deleted_at IS NULL
	`

	_, err := r.DB.ExecContext(ctx, query, userID)

	return err
}
