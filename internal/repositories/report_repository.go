package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"pos-saas-backend/internal/models"
	"pos-saas-backend/internal/requests"
)

type ReportRepository struct {
	DB *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{
		DB: db,
	}
}

func (r *ReportRepository) GetDailySalesReport(ctx context.Context, request requests.DailySalesReportRequest) (*models.DailySalesReport, []models.DailySalesPaymentBreakdown, []models.DailySalesTopProduct, error) {
	whereClauses := []string{
		"t.deleted_at IS NULL",
		"t.status = 'paid'",
		"DATE(t.transaction_date AT TIME ZONE 'Asia/Jakarta') = $1::DATE",
	}

	args := []interface{}{
		strings.TrimSpace(request.Date),
	}

	argPosition := 2

	if request.StoreID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("t.store_id = $%d", argPosition))
		args = append(args, request.StoreID)
		argPosition++
	}

	if request.BranchID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("t.branch_id = $%d", argPosition))
		args = append(args, request.BranchID)
		argPosition++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	summaryQuery := fmt.Sprintf(`
		SELECT
			$1::TEXT AS report_date,
			COALESCE(MAX(t.store_id), 0) AS store_id,
			COALESCE(MAX(t.branch_id), 0) AS branch_id,
			COUNT(DISTINCT t.id) AS total_transactions,
			COALESCE(SUM(ti.qty), 0) AS total_items_sold,
			COALESCE(MAX(summary.subtotal), 0) AS subtotal,
			COALESCE(MAX(summary.discount_total), 0) AS discount_total,
			COALESCE(MAX(summary.grand_total), 0) AS grand_total,
			COALESCE(MAX(summary.cash_total), 0) AS cash_total,
			COALESCE(MAX(summary.transfer_total), 0) AS transfer_total,
			COALESCE(MAX(summary.change_total), 0) AS change_total
		FROM transactions t
		LEFT JOIN transaction_items ti ON ti.transaction_id = t.id
		CROSS JOIN (
			SELECT
				COALESCE(SUM(t2.subtotal), 0) AS subtotal,
				COALESCE(SUM(t2.discount_total), 0) AS discount_total,
				COALESCE(SUM(t2.grand_total), 0) AS grand_total,
				COALESCE(SUM(t2.cash_amount), 0) AS cash_total,
				COALESCE(SUM(t2.transfer_amount), 0) AS transfer_total,
				COALESCE(SUM(t2.change_amount), 0) AS change_total
			FROM transactions t2
			WHERE %s
		) summary
		WHERE %s
	`, replaceTransactionAlias(whereSQL, "t", "t2"), whereSQL)

	var report models.DailySalesReport

	err := r.DB.QueryRowContext(ctx, summaryQuery, args...).Scan(
		&report.Date,
		&report.StoreID,
		&report.BranchID,
		&report.TotalTransactions,
		&report.TotalItemsSold,
		&report.Subtotal,
		&report.DiscountTotal,
		&report.GrandTotal,
		&report.CashTotal,
		&report.TransferTotal,
		&report.ChangeTotal,
	)

	if err != nil {
		return nil, nil, nil, err
	}

	paymentBreakdown, err := r.getDailySalesPaymentBreakdown(ctx, whereSQL, args)
	if err != nil {
		return nil, nil, nil, err
	}

	topProducts, err := r.getDailySalesTopProducts(ctx, whereSQL, args)
	if err != nil {
		return nil, nil, nil, err
	}

	return &report, paymentBreakdown, topProducts, nil
}

func (r *ReportRepository) getDailySalesPaymentBreakdown(ctx context.Context, whereSQL string, args []interface{}) ([]models.DailySalesPaymentBreakdown, error) {
	query := fmt.Sprintf(`
		SELECT
			t.payment_method,
			COUNT(t.id) AS total_transactions,
			COALESCE(SUM(t.grand_total), 0) AS grand_total,
			COALESCE(SUM(t.cash_amount), 0) AS cash_total,
			COALESCE(SUM(t.transfer_amount), 0) AS transfer_total
		FROM transactions t
		WHERE %s
		GROUP BY t.payment_method
		ORDER BY t.payment_method ASC
	`, whereSQL)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	breakdowns := make([]models.DailySalesPaymentBreakdown, 0)

	for rows.Next() {
		var breakdown models.DailySalesPaymentBreakdown

		err := rows.Scan(
			&breakdown.PaymentMethod,
			&breakdown.TotalTransactions,
			&breakdown.GrandTotal,
			&breakdown.CashTotal,
			&breakdown.TransferTotal,
		)

		if err != nil {
			return nil, err
		}

		breakdowns = append(breakdowns, breakdown)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return breakdowns, nil
}

func (r *ReportRepository) getDailySalesTopProducts(ctx context.Context, whereSQL string, args []interface{}) ([]models.DailySalesTopProduct, error) {
	query := fmt.Sprintf(`
		SELECT
			ti.product_id,
			ti.product_name_snapshot,
			ti.product_sku_snapshot,
			COALESCE(SUM(ti.qty), 0) AS total_qty,
			COALESCE(SUM(ti.subtotal), 0) AS total_sales
		FROM transactions t
		INNER JOIN transaction_items ti ON ti.transaction_id = t.id
		WHERE %s
		AND ti.deleted_at IS NULL
		GROUP BY
			ti.product_id,
			ti.product_name_snapshot,
			ti.product_sku_snapshot
		ORDER BY total_qty DESC, total_sales DESC
		LIMIT 10
	`, whereSQL)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	topProducts := make([]models.DailySalesTopProduct, 0)

	for rows.Next() {
		var product models.DailySalesTopProduct

		err := rows.Scan(
			&product.ProductID,
			&product.ProductName,
			&product.ProductSKU,
			&product.TotalQty,
			&product.TotalSales,
		)

		if err != nil {
			return nil, err
		}

		topProducts = append(topProducts, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return topProducts, nil
}

func replaceTransactionAlias(whereSQL string, oldAlias string, newAlias string) string {
	replaced := strings.ReplaceAll(whereSQL, oldAlias+".", newAlias+".")

	return replaced
}
