package repositories

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/models"
	"pos-saas-backend/internal/requests"
)

type TransactionRepository struct {
	DB *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{
		DB: db,
	}
}

type transactionProductData struct {
	ID           int64
	Name         string
	SKU          string
	SellingPrice float64
	Stock        int
}

func (r *TransactionRepository) FindAll(ctx context.Context, filter requests.TransactionFilterRequest) ([]models.Transaction, int64, error) {
	whereClauses := []string{
		"t.deleted_at IS NULL",
	}

	args := make([]interface{}, 0)
	argPosition := 1

	if filter.StoreID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("t.store_id = $%d", argPosition))
		args = append(args, filter.StoreID)
		argPosition++
	}

	if filter.BranchID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("t.branch_id = $%d", argPosition))
		args = append(args, filter.BranchID)
		argPosition++
	}

	if filter.CashierID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("t.cashier_id = $%d", argPosition))
		args = append(args, filter.CashierID)
		argPosition++
	}

	if strings.TrimSpace(filter.Date) != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("DATE(t.transaction_date AT TIME ZONE 'Asia/Jakarta') = $%d", argPosition))
		args = append(args, strings.TrimSpace(filter.Date))
		argPosition++
	}

	if strings.TrimSpace(filter.PaymentMethod) != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("t.payment_method = $%d", argPosition))
		args = append(args, strings.ToLower(strings.TrimSpace(filter.PaymentMethod)))
		argPosition++
	}

	if strings.TrimSpace(filter.Search) != "" {
		whereClauses = append(
			whereClauses,
			fmt.Sprintf("(t.transaction_number ILIKE $%d OR t.customer_name ILIKE $%d)", argPosition, argPosition),
		)

		args = append(args, "%"+strings.TrimSpace(filter.Search)+"%")
		argPosition++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM transactions t
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
			t.id,

			t.store_id,
			s.name AS store_name,

			t.branch_id,
			b.name AS branch_name,

			t.cashier_id,
			u.name AS cashier_name,

			t.transaction_number,
			t.customer_name,

			t.payment_method,

			t.subtotal,
			t.discount_total,
			t.grand_total,

			t.cash_amount,
			t.transfer_amount,
			t.change_amount,

			t.notes,
			t.status,

			t.transaction_date,
			t.created_at,
			t.updated_at
		FROM transactions t
		INNER JOIN stores s ON s.id = t.store_id
		INNER JOIN branches b ON b.id = t.branch_id
		INNER JOIN users u ON u.id = t.cashier_id
		WHERE %s
		ORDER BY t.id DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argPosition, argPosition+1)

	rows, err := r.DB.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	transactions := make([]models.Transaction, 0)

	for rows.Next() {
		transaction, err := scanTransaction(rows)
		if err != nil {
			return nil, 0, err
		}

		transactions = append(transactions, *transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *TransactionRepository) FindByID(ctx context.Context, id int64) (*models.Transaction, error) {
	query := `
		SELECT
			t.id,

			t.store_id,
			s.name AS store_name,

			t.branch_id,
			b.name AS branch_name,

			t.cashier_id,
			u.name AS cashier_name,

			t.transaction_number,
			t.customer_name,

			t.payment_method,

			t.subtotal,
			t.discount_total,
			t.grand_total,

			t.cash_amount,
			t.transfer_amount,
			t.change_amount,

			t.notes,
			t.status,

			t.transaction_date,
			t.created_at,
			t.updated_at
		FROM transactions t
		INNER JOIN stores s ON s.id = t.store_id
		INNER JOIN branches b ON b.id = t.branch_id
		INNER JOIN users u ON u.id = t.cashier_id
		WHERE t.id = $1
		AND t.deleted_at IS NULL
		LIMIT 1
	`

	row := r.DB.QueryRowContext(ctx, query, id)

	transaction, err := scanTransactionRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	items, err := r.FindItemsByTransactionID(ctx, id)
	if err != nil {
		return nil, err
	}

	transaction.Items = items

	return transaction, nil
}

func (r *TransactionRepository) FindItemsByTransactionID(ctx context.Context, transactionID int64) ([]models.TransactionItem, error) {
	query := `
		SELECT
			id,
			transaction_id,
			product_id,
			product_name_snapshot,
			product_sku_snapshot,
			qty,
			price,
			discount,
			subtotal,
			created_at,
			updated_at
		FROM transaction_items
		WHERE transaction_id = $1
		AND deleted_at IS NULL
		ORDER BY id ASC
	`

	rows, err := r.DB.QueryContext(ctx, query, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.TransactionItem, 0)

	for rows.Next() {
		var item models.TransactionItem

		err := rows.Scan(
			&item.ID,
			&item.TransactionID,
			&item.ProductID,
			&item.ProductNameSnapshot,
			&item.ProductSKUSnapshot,
			&item.Qty,
			&item.Price,
			&item.Discount,
			&item.Subtotal,
			&item.CreatedAt,
			&item.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *TransactionRepository) Create(ctx context.Context, cashierID int64, request requests.TransactionRequest) (*models.Transaction, error) {
	paymentMethod := strings.ToLower(strings.TrimSpace(request.PaymentMethod))

	if paymentMethod != "cash" && paymentMethod != "transfer" && paymentMethod != "mixed" {
		return nil, errors.New("payment_method must be cash, transfer, or mixed")
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	transactionNumber, err := generateTransactionNumber()
	if err != nil {
		return nil, err
	}

	subtotal := 0.0
	discountTotal := 0.0

	type preparedItem struct {
		Product  transactionProductData
		Qty      int
		Discount float64
		Subtotal float64
	}

	preparedItems := make([]preparedItem, 0)

	for _, item := range request.Items {
		if item.Qty <= 0 {
			err = errors.New("item qty must be greater than zero")
			return nil, err
		}

		if item.Discount < 0 {
			err = errors.New("item discount cannot be negative")
			return nil, err
		}

		product, productErr := r.findProductForUpdate(ctx, tx, item.ProductID, request.StoreID, request.BranchID)
		if productErr != nil {
			err = productErr
			return nil, err
		}

		if product == nil {
			err = errors.New("product not found or inactive")
			return nil, err
		}

		if product.Stock < item.Qty {
			err = fmt.Errorf("insufficient stock for product %s", product.Name)
			return nil, err
		}

		grossItemSubtotal := product.SellingPrice * float64(item.Qty)
		itemSubtotal := grossItemSubtotal - item.Discount

		if itemSubtotal < 0 {
			err = fmt.Errorf("discount cannot be greater than subtotal for product %s", product.Name)
			return nil, err
		}

		subtotal += grossItemSubtotal
		discountTotal += item.Discount

		preparedItems = append(preparedItems, preparedItem{
			Product:  *product,
			Qty:      item.Qty,
			Discount: item.Discount,
			Subtotal: itemSubtotal,
		})
	}

	grandTotal := subtotal - discountTotal
	totalPaid := request.CashAmount + request.TransferAmount

	if totalPaid < grandTotal {
		err = errors.New("payment amount is less than grand total")
		return nil, err
	}

	changeAmount := totalPaid - grandTotal

	insertTransactionQuery := `
		INSERT INTO transactions (
			store_id,
			branch_id,
			cashier_id,
			transaction_number,
			customer_name,
			payment_method,
			subtotal,
			discount_total,
			grand_total,
			cash_amount,
			transfer_amount,
			change_amount,
			notes,
			status
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			NULLIF($5, ''),
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			NULLIF($13, ''),
			'paid'
		)
		RETURNING id
	`

	var transactionID int64

	err = tx.QueryRowContext(
		ctx,
		insertTransactionQuery,
		request.StoreID,
		request.BranchID,
		cashierID,
		transactionNumber,
		request.CustomerName,
		paymentMethod,
		subtotal,
		discountTotal,
		grandTotal,
		request.CashAmount,
		request.TransferAmount,
		changeAmount,
		request.Notes,
	).Scan(&transactionID)

	if err != nil {
		return nil, err
	}

	insertItemQuery := `
		INSERT INTO transaction_items (
			transaction_id,
			product_id,
			product_name_snapshot,
			product_sku_snapshot,
			qty,
			price,
			discount,
			subtotal
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8
		)
	`

	updateStockQuery := `
		UPDATE products
		SET
			stock = stock - $1,
			updated_at = NOW()
		WHERE id = $2
		AND deleted_at IS NULL
	`

	for _, item := range preparedItems {
		_, err = tx.ExecContext(
			ctx,
			insertItemQuery,
			transactionID,
			item.Product.ID,
			item.Product.Name,
			item.Product.SKU,
			item.Qty,
			item.Product.SellingPrice,
			item.Discount,
			item.Subtotal,
		)

		if err != nil {
			return nil, err
		}

		_, err = tx.ExecContext(
			ctx,
			updateStockQuery,
			item.Qty,
			item.Product.ID,
		)

		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.FindByID(ctx, transactionID)
}

func (r *TransactionRepository) findProductForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	productID int64,
	storeID int64,
	branchID int64,
) (*transactionProductData, error) {
	query := `
		SELECT
			id,
			name,
			sku,
			selling_price,
			stock
		FROM products
		WHERE id = $1
		AND store_id = $2
		AND branch_id = $3
		AND is_active = TRUE
		AND deleted_at IS NULL
		FOR UPDATE
	`

	var product transactionProductData

	err := tx.QueryRowContext(ctx, query, productID, storeID, branchID).Scan(
		&product.ID,
		&product.Name,
		&product.SKU,
		&product.SellingPrice,
		&product.Stock,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &product, nil
}

type transactionScanner interface {
	Scan(dest ...interface{}) error
}

func scanTransactionRow(scanner transactionScanner) (*models.Transaction, error) {
	var transaction models.Transaction

	var customerName sql.NullString
	var notes sql.NullString

	err := scanner.Scan(
		&transaction.ID,

		&transaction.StoreID,
		&transaction.StoreName,

		&transaction.BranchID,
		&transaction.BranchName,

		&transaction.CashierID,
		&transaction.CashierName,

		&transaction.TransactionNumber,
		&customerName,

		&transaction.PaymentMethod,

		&transaction.Subtotal,
		&transaction.DiscountTotal,
		&transaction.GrandTotal,

		&transaction.CashAmount,
		&transaction.TransferAmount,
		&transaction.ChangeAmount,

		&notes,
		&transaction.Status,

		&transaction.TransactionDate,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	transaction.CustomerName = helpers.NullableString(customerName)
	transaction.Notes = helpers.NullableString(notes)

	return &transaction, nil
}

func scanTransaction(rows *sql.Rows) (*models.Transaction, error) {
	return scanTransactionRow(rows)
}

func generateTransactionNumber() (string, error) {
	randomBytes := make([]byte, 4)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	randomPart := strings.ToUpper(hex.EncodeToString(randomBytes))
	datePart := time.Now().Format("20060102150405")

	return fmt.Sprintf("TRX-%s-%s", datePart, randomPart), nil
}
