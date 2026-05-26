package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"pos-saas-backend/internal/models"
	"pos-saas-backend/internal/requests"
)

type DashboardRepository struct {
	DB *sql.DB
}

func NewDashboardRepository(db *sql.DB) *DashboardRepository {
	return &DashboardRepository{
		DB: db,
	}
}

func (r *DashboardRepository) GetDashboard(ctx context.Context, request requests.DashboardRequest) (*models.DashboardResponse, error) {
	summary, err := r.GetSummary(ctx, request)
	if err != nil {
		return nil, err
	}

	salesPerHour, err := r.GetSalesPerHour(ctx, request)
	if err != nil {
		return nil, err
	}

	lowStockProducts, err := r.GetLowStockProducts(ctx, request)
	if err != nil {
		return nil, err
	}

	recentTransactions, err := r.GetRecentTransactions(ctx, request)
	if err != nil {
		return nil, err
	}

	return &models.DashboardResponse{
		Summary:            *summary,
		SalesPerHour:       salesPerHour,
		LowStockProducts:   lowStockProducts,
		RecentTransactions: recentTransactions,
	}, nil
}

func (r *DashboardRepository) GetSummary(ctx context.Context, request requests.DashboardRequest) (*models.DashboardSummary, error) {
	whereSQL, args := buildDashboardTransactionWhere(request, "t")

	summaryQuery := fmt.Sprintf(`
		SELECT
			$1::TEXT AS report_date,
			COALESCE(MAX(t.store_id), 0) AS store_id,
			COALESCE(MAX(t.branch_id), 0) AS branch_id,
			COUNT(t.id) AS total_transactions,
			COALESCE(SUM(t.grand_total), 0) AS total_sales,
			COALESCE(SUM(t.cash_amount), 0) AS total_cash,
			COALESCE(SUM(t.transfer_amount), 0) AS total_transfer,
			COALESCE(SUM(t.discount_total), 0) AS total_discount
		FROM transactions t
		WHERE %s
	`, whereSQL)

	var summary models.DashboardSummary

	err := r.DB.QueryRowContext(ctx, summaryQuery, args...).Scan(
		&summary.Date,
		&summary.StoreID,
		&summary.BranchID,
		&summary.TotalTransactions,
		&summary.TotalSales,
		&summary.TotalCash,
		&summary.TotalTransfer,
		&summary.TotalDiscount,
	)

	if err != nil {
		return nil, err
	}

	totalItemsSold, err := r.GetTotalItemsSold(ctx, request)
	if err != nil {
		return nil, err
	}

	lowStockCount, err := r.GetLowStockCount(ctx, request)
	if err != nil {
		return nil, err
	}

	summary.TotalItemsSold = totalItemsSold
	summary.LowStockCount = lowStockCount

	if request.StoreID > 0 {
		summary.StoreID = request.StoreID
	}

	if request.BranchID > 0 {
		summary.BranchID = request.BranchID
	}

	return &summary, nil
}

func (r *DashboardRepository) GetTotalItemsSold(ctx context.Context, request requests.DashboardRequest) (int64, error) {
	whereSQL, args := buildDashboardTransactionWhere(request, "t")

	query := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(ti.qty), 0) AS total_items_sold
		FROM transactions t
		INNER JOIN transaction_items ti ON ti.transaction_id = t.id
		WHERE %s
		AND ti.deleted_at IS NULL
	`, whereSQL)

	var totalItemsSold int64

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(&totalItemsSold)
	if err != nil {
		return 0, err
	}

	return totalItemsSold, nil
}

func (r *DashboardRepository) GetLowStockCount(ctx context.Context, request requests.DashboardRequest) (int64, error) {
	whereSQL, args := buildDashboardProductWhere(request, "p")

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM products p
		WHERE %s
	`, whereSQL)

	var total int64

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *DashboardRepository) GetSalesPerHour(ctx context.Context, request requests.DashboardRequest) ([]models.DashboardHourlySales, error) {
	joinSQL, args := buildDashboardTransactionWhere(request, "t")

	query := fmt.Sprintf(`
		SELECT
			h.hour,
			LPAD(h.hour::TEXT, 2, '0') || ':00' AS label,
			COUNT(t.id) AS total_transactions,
			COALESCE(SUM(t.grand_total), 0) AS total_sales
		FROM generate_series(0, 23) AS h(hour)
		LEFT JOIN transactions t
			ON EXTRACT(HOUR FROM t.transaction_date AT TIME ZONE 'Asia/Jakarta')::INT = h.hour
			AND %s
		GROUP BY h.hour
		ORDER BY h.hour ASC
	`, joinSQL)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hourlySales := make([]models.DashboardHourlySales, 0)

	for rows.Next() {
		var item models.DashboardHourlySales

		err := rows.Scan(
			&item.Hour,
			&item.Label,
			&item.TotalTransactions,
			&item.TotalSales,
		)

		if err != nil {
			return nil, err
		}

		hourlySales = append(hourlySales, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hourlySales, nil
}

func (r *DashboardRepository) GetLowStockProducts(ctx context.Context, request requests.DashboardRequest) ([]models.Product, error) {
	whereSQL, args := buildDashboardProductWhere(request, "p")

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
		ORDER BY p.stock ASC, p.id DESC
		LIMIT 10
	`, whereSQL)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]models.Product, 0)

	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}

		products = append(products, *product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *DashboardRepository) GetRecentTransactions(ctx context.Context, request requests.DashboardRequest) ([]models.Transaction, error) {
	whereSQL, args := buildDashboardTransactionWhere(request, "t")

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

			t.void_reason,
			t.voided_at,
			t.voided_by,
			vu.name AS voided_by_name,

			t.transaction_date,
			t.created_at,
			t.updated_at
		FROM transactions t
		INNER JOIN stores s ON s.id = t.store_id
		INNER JOIN branches b ON b.id = t.branch_id
		INNER JOIN users u ON u.id = t.cashier_id
		LEFT JOIN users vu ON vu.id = t.voided_by
		WHERE %s
		ORDER BY t.id DESC
		LIMIT 5
	`, whereSQL)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]models.Transaction, 0)

	for rows.Next() {
		transaction, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, *transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func buildDashboardTransactionWhere(request requests.DashboardRequest, alias string) (string, []interface{}) {
	whereClauses := []string{
		fmt.Sprintf("%s.deleted_at IS NULL", alias),
		fmt.Sprintf("%s.status = 'paid'", alias),
		fmt.Sprintf("DATE(%s.transaction_date AT TIME ZONE 'Asia/Jakarta') = $1::DATE", alias),
	}

	args := []interface{}{
		strings.TrimSpace(request.Date),
	}

	argPosition := 2

	if request.StoreID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("%s.store_id = $%d", alias, argPosition))
		args = append(args, request.StoreID)
		argPosition++
	}

	if request.BranchID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("%s.branch_id = $%d", alias, argPosition))
		args = append(args, request.BranchID)
		argPosition++
	}

	return strings.Join(whereClauses, " AND "), args
}

func buildDashboardProductWhere(request requests.DashboardRequest, alias string) (string, []interface{}) {
	whereClauses := []string{
		fmt.Sprintf("%s.deleted_at IS NULL", alias),
		fmt.Sprintf("%s.is_active = TRUE", alias),
		fmt.Sprintf("%s.stock <= %s.min_stock", alias, alias),
	}

	args := make([]interface{}, 0)
	argPosition := 1

	if request.StoreID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("%s.store_id = $%d", alias, argPosition))
		args = append(args, request.StoreID)
		argPosition++
	}

	if request.BranchID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("%s.branch_id = $%d", alias, argPosition))
		args = append(args, request.BranchID)
		argPosition++
	}

	return strings.Join(whereClauses, " AND "), args
}
