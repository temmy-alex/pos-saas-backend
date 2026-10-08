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

type CustomerRepository struct {
	DB *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{DB: db}
}

func (r *CustomerRepository) FindAll(ctx context.Context, filter requests.CustomerFilterRequest) ([]models.Customer, int64, error) {
	whereClauses := []string{"c.deleted_at IS NULL"}
	args := make([]interface{}, 0)
	argPosition := 1

	if filter.StoreID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("c.store_id = $%d", argPosition))
		args = append(args, filter.StoreID)
		argPosition++
	}

	if filter.BranchID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("c.branch_id = $%d", argPosition))
		args = append(args, filter.BranchID)
		argPosition++
	}

	if search := strings.TrimSpace(filter.Search); search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(c.name ILIKE $%d OR c.phone ILIKE $%d OR c.email ILIKE $%d)",
			argPosition,
			argPosition,
			argPosition,
		))
		args = append(args, "%"+search+"%")
		argPosition++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM customers c WHERE %s", whereSQL)
	if err := r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	queryArgs := append(args, filter.Limit, offset)

	query := fmt.Sprintf(`
		SELECT
			c.id,
			c.store_id,
			s.name AS store_name,
			c.branch_id,
			b.name AS branch_name,
			c.name,
			c.phone,
			c.email,
			0 AS visit_count,
			c.created_at,
			c.updated_at
		FROM customers c
		INNER JOIN stores s ON s.id = c.store_id
		INNER JOIN branches b ON b.id = c.branch_id
		WHERE %s
		ORDER BY visit_count DESC, c.id DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argPosition, argPosition+1)

	rows, err := r.DB.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	customers := make([]models.Customer, 0)
	for rows.Next() {
		customer, err := scanCustomer(rows)
		if err != nil {
			return nil, 0, err
		}
		customers = append(customers, *customer)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

func (r *CustomerRepository) FindByID(ctx context.Context, id int64) (*models.Customer, error) {
	query := `
		SELECT
			c.id,
			c.store_id,
			s.name AS store_name,
			c.branch_id,
			b.name AS branch_name,
			c.name,
			c.phone,
			c.email,
			0 AS visit_count,
			c.created_at,
			c.updated_at
		FROM customers c
		INNER JOIN stores s ON s.id = c.store_id
		INNER JOIN branches b ON b.id = c.branch_id
		WHERE c.id = $1
		AND c.deleted_at IS NULL
		LIMIT 1
	`

	customer, err := scanCustomer(r.DB.QueryRowContext(ctx, query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return customer, nil
}

func (r *CustomerRepository) Create(ctx context.Context, storeID, branchID int64, request requests.CustomerRequest) (*models.Customer, error) {
	query := `
		INSERT INTO customers (store_id, branch_id, name, phone, email)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''))
		RETURNING id
	`

	var customerID int64
	if err := r.DB.QueryRowContext(
		ctx,
		query,
		storeID,
		branchID,
		strings.TrimSpace(request.Name),
		strings.TrimSpace(request.Phone),
		strings.TrimSpace(request.Email),
	).Scan(&customerID); err != nil {
		return nil, err
	}

	return r.FindByID(ctx, customerID)
}

func (r *CustomerRepository) Update(ctx context.Context, id, storeID, branchID int64, request requests.CustomerRequest) (*models.Customer, error) {
	query := `
		UPDATE customers
		SET
			store_id = $1,
			branch_id = $2,
			name = $3,
			phone = $4,
			email = NULLIF($5, ''),
			updated_at = NOW()
		WHERE id = $6
		AND deleted_at IS NULL
		RETURNING id
	`

	var customerID int64
	err := r.DB.QueryRowContext(
		ctx,
		query,
		storeID,
		branchID,
		strings.TrimSpace(request.Name),
		strings.TrimSpace(request.Phone),
		strings.TrimSpace(request.Email),
		id,
	).Scan(&customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return r.FindByID(ctx, customerID)
}

func (r *CustomerRepository) Delete(ctx context.Context, id int64) (bool, error) {
	result, err := r.DB.ExecContext(ctx, `
		UPDATE customers
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

type customerScanner interface {
	Scan(dest ...interface{}) error
}

func scanCustomer(scanner customerScanner) (*models.Customer, error) {
	var customer models.Customer
	var email sql.NullString

	err := scanner.Scan(
		&customer.ID,
		&customer.StoreID,
		&customer.StoreName,
		&customer.BranchID,
		&customer.BranchName,
		&customer.Name,
		&customer.Phone,
		&email,
		&customer.VisitCount,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	customer.Email = helpers.NullableString(email)
	return &customer, nil
}
