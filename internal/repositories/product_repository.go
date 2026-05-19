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

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		DB: db,
	}
}

func (r *ProductRepository) FindAll(ctx context.Context, filter requests.ProductFilterRequest) ([]models.Product, int64, error) {
	whereClauses := []string{
		"p.deleted_at IS NULL",
	}

	args := make([]interface{}, 0)
	argPosition := 1

	if filter.StoreID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("p.store_id = $%d", argPosition))
		args = append(args, filter.StoreID)
		argPosition++
	}

	if filter.BranchID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("p.branch_id = $%d", argPosition))
		args = append(args, filter.BranchID)
		argPosition++
	}

	if filter.CategoryID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("p.category_id = $%d", argPosition))
		args = append(args, filter.CategoryID)
		argPosition++
	}

	if filter.IsActive != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("p.is_active = $%d", argPosition))
		args = append(args, *filter.IsActive)
		argPosition++
	}

	if strings.TrimSpace(filter.Search) != "" {
		whereClauses = append(
			whereClauses,
			fmt.Sprintf("(p.name ILIKE $%d OR p.sku ILIKE $%d OR p.barcode ILIKE $%d)", argPosition, argPosition, argPosition),
		)

		args = append(args, "%"+strings.TrimSpace(filter.Search)+"%")
		argPosition++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM products p
		WHERE %s
	`, whereSQL)

	var total int64

	if err := r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	queryArgs := append(args, limit, offset)

	query := fmt.Sprintf(`
		SELECT
			p.id,

			p.store_id,
			s.name AS store_name,

			p.branch_id,
			b.name AS branch_name,

			p.category_id,
			c.name AS category_name,

			p.name,
			p.sku,
			p.barcode,
			p.description,

			p.cost_price,
			p.selling_price,

			p.stock,
			p.min_stock,

			p.image_path,
			p.image_url,
			p.image_disk,

			p.is_active,
			p.created_at,
			p.updated_at
		FROM products p
		INNER JOIN stores s ON s.id = p.store_id
		INNER JOIN branches b ON b.id = p.branch_id
		INNER JOIN categories c ON c.id = p.category_id
		WHERE %s
		ORDER BY p.id DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argPosition, argPosition+1)

	rows, err := r.DB.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := make([]models.Product, 0)

	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}

		products = append(products, *product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id int64) (*models.Product, error) {
	query := `
		SELECT
			p.id,

			p.store_id,
			s.name AS store_name,

			p.branch_id,
			b.name AS branch_name,

			p.category_id,
			c.name AS category_name,

			p.name,
			p.sku,
			p.barcode,
			p.description,

			p.cost_price,
			p.selling_price,

			p.stock,
			p.min_stock,

			p.image_path,
			p.image_url,
			p.image_disk,

			p.is_active,
			p.created_at,
			p.updated_at
		FROM products p
		INNER JOIN stores s ON s.id = p.store_id
		INNER JOIN branches b ON b.id = p.branch_id
		INNER JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1
		AND p.deleted_at IS NULL
		LIMIT 1
	`

	row := r.DB.QueryRowContext(ctx, query, id)

	product, err := scanProductRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return product, nil
}

func (r *ProductRepository) Create(ctx context.Context, request requests.ProductRequest) (*models.Product, error) {
	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	query := `
		INSERT INTO products (
			store_id,
			branch_id,
			category_id,

			name,
			sku,
			barcode,
			description,

			cost_price,
			selling_price,

			stock,
			min_stock,

			image_path,
			image_url,
			image_disk,

			is_active
		) VALUES (
			$1,
			$2,
			$3,

			$4,
			$5,
			NULLIF($6, ''),
			NULLIF($7, ''),

			$8,
			$9,

			$10,
			$11,

			$12,
			$13,
			$14,

			$15
		)
		RETURNING id
	`

	var productID int64

	err := r.DB.QueryRowContext(
		ctx,
		query,
		request.StoreID,
		request.BranchID,
		request.CategoryID,

		request.Name,
		request.SKU,
		request.Barcode,
		request.Description,

		request.CostPrice,
		request.SellingPrice,

		request.Stock,
		request.MinStock,

		request.ImagePath,
		request.ImageURL,
		request.ImageDisk,

		isActive,
	).Scan(&productID)

	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, productID)
}

func (r *ProductRepository) Update(ctx context.Context, id int64, request requests.ProductRequest) (*models.Product, error) {
	currentProduct, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if currentProduct == nil {
		return nil, nil
	}

	isActive := currentProduct.IsActive
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	imagePath := currentProduct.ImagePath
	imageURL := currentProduct.ImageURL
	imageDisk := currentProduct.ImageDisk

	if request.ImagePath != nil {
		imagePath = request.ImagePath
	}

	if request.ImageURL != nil {
		imageURL = request.ImageURL
	}

	if request.ImageDisk != nil {
		imageDisk = request.ImageDisk
	}

	query := `
		UPDATE products
		SET
			store_id = $1,
			branch_id = $2,
			category_id = $3,

			name = $4,
			sku = $5,
			barcode = NULLIF($6, ''),
			description = NULLIF($7, ''),

			cost_price = $8,
			selling_price = $9,

			stock = $10,
			min_stock = $11,

			image_path = $12,
			image_url = $13,
			image_disk = $14,

			is_active = $15,
			updated_at = NOW()
		WHERE id = $16
		AND deleted_at IS NULL
		RETURNING id
	`

	var productID int64

	err = r.DB.QueryRowContext(
		ctx,
		query,
		request.StoreID,
		request.BranchID,
		request.CategoryID,

		request.Name,
		request.SKU,
		request.Barcode,
		request.Description,

		request.CostPrice,
		request.SellingPrice,

		request.Stock,
		request.MinStock,

		imagePath,
		imageURL,
		imageDisk,

		isActive,
		id,
	).Scan(&productID)

	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, productID)
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) (bool, error) {
	query := `
		UPDATE products
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

type productScanner interface {
	Scan(dest ...interface{}) error
}

func scanProductRow(scanner productScanner) (*models.Product, error) {
	var product models.Product

	var barcode sql.NullString
	var description sql.NullString
	var imagePath sql.NullString
	var imageURL sql.NullString
	var imageDisk sql.NullString

	err := scanner.Scan(
		&product.ID,

		&product.StoreID,
		&product.StoreName,

		&product.BranchID,
		&product.BranchName,

		&product.CategoryID,
		&product.CategoryName,

		&product.Name,
		&product.SKU,
		&barcode,
		&description,

		&product.CostPrice,
		&product.SellingPrice,

		&product.Stock,
		&product.MinStock,

		&imagePath,
		&imageURL,
		&imageDisk,

		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	product.Barcode = helpers.NullableString(barcode)
	product.Description = helpers.NullableString(description)
	product.ImagePath = helpers.NullableString(imagePath)
	product.ImageURL = helpers.NullableString(imageURL)
	product.ImageDisk = helpers.NullableString(imageDisk)

	return &product, nil
}

func scanProduct(rows *sql.Rows) (*models.Product, error) {
	return scanProductRow(rows)
}
