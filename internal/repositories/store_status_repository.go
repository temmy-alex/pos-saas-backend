package repositories

import (
	"context"
	"database/sql"
	"errors"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/models"
)

var (
	ErrStoreAlreadyOpen   = errors.New("store is already open for this business date")
	ErrStoreAlreadyClosed = errors.New("store is already closed for this business date")
	ErrStoreNotOpen       = errors.New("store must be opened before it can be closed")
)

type StoreStatusRepository struct {
	DB *sql.DB
}

func NewStoreStatusRepository(db *sql.DB) *StoreStatusRepository {
	return &StoreStatusRepository{DB: db}
}

func (r *StoreStatusRepository) GetOrCreate(
	ctx context.Context,
	storeID, branchID int64,
	businessDate string,
) (*models.StoreStatus, models.StoreSalesSummary, error) {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO store_statuses (
			business_date, store_id, branch_id, is_open
		) VALUES ($1::DATE, $2, $3, FALSE)
		ON CONFLICT (business_date, branch_id) DO NOTHING
	`, businessDate, storeID, branchID)
	if err != nil {
		return nil, models.StoreSalesSummary{}, err
	}

	status, err := r.find(ctx, r.DB, branchID, businessDate, false)
	if err != nil {
		return nil, models.StoreSalesSummary{}, err
	}

	summary, err := r.salesSummary(ctx, r.DB, branchID, businessDate)
	if err != nil {
		return nil, models.StoreSalesSummary{}, err
	}

	if status.CashOpen != nil {
		summary.ExpectedCash += *status.CashOpen
	}

	return status, summary, nil
}

func (r *StoreStatusRepository) Open(
	ctx context.Context,
	storeID, branchID, userID int64,
	businessDate string,
	cashOpen float64,
	scheduledCloseTime *string,
) (*models.StoreStatus, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO store_statuses (business_date, store_id, branch_id, is_open)
		VALUES ($1::DATE, $2, $3, FALSE)
		ON CONFLICT (business_date, branch_id) DO NOTHING
	`, businessDate, storeID, branchID)
	if err != nil {
		return nil, err
	}

	status, err := r.find(ctx, tx, branchID, businessDate, true)
	if err != nil {
		return nil, err
	}
	if status.IsOpen {
		return nil, ErrStoreAlreadyOpen
	}
	if status.OpenedAt != nil || status.ClosedAtTimestamp != nil {
		return nil, ErrStoreAlreadyClosed
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE store_statuses
		SET
			store_id = $1,
			is_open = TRUE,
			cash_open = $2,
			cash_close = NULL,
			expected_cash_at_close = NULL,
			cash_difference = NULL,
			scheduled_close_time = NULLIF($3, '')::TIME,
			opened_by = $4,
			closed_by = NULL,
			opened_at = NOW(),
			closed_at = NULL,
			updated_at = NOW()
		WHERE id = $5
	`, storeID, cashOpen, nullableStringValue(scheduledCloseTime), userID, status.ID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.find(ctx, r.DB, branchID, businessDate, false)
}

func (r *StoreStatusRepository) Close(
	ctx context.Context,
	branchID, userID int64,
	businessDate string,
	cashClose float64,
) (*models.StoreStatus, models.StoreSalesSummary, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, models.StoreSalesSummary{}, err
	}
	defer func() { _ = tx.Rollback() }()

	status, err := r.find(ctx, tx, branchID, businessDate, true)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.StoreSalesSummary{}, ErrStoreNotOpen
		}
		return nil, models.StoreSalesSummary{}, err
	}
	if status.OpenedAt == nil {
		return nil, models.StoreSalesSummary{}, ErrStoreNotOpen
	}
	if !status.IsOpen {
		return nil, models.StoreSalesSummary{}, ErrStoreAlreadyClosed
	}

	summary, err := r.salesSummary(ctx, tx, branchID, businessDate)
	if err != nil {
		return nil, models.StoreSalesSummary{}, err
	}
	if status.CashOpen != nil {
		summary.ExpectedCash += *status.CashOpen
	}
	cashDifference := cashClose - summary.ExpectedCash

	_, err = tx.ExecContext(ctx, `
		UPDATE store_statuses
		SET
			is_open = FALSE,
			cash_close = $1,
			expected_cash_at_close = $2,
			cash_difference = $3,
			closed_by = $4,
			closed_at = NOW(),
			updated_at = NOW()
		WHERE id = $5
	`, cashClose, summary.ExpectedCash, cashDifference, userID, status.ID)
	if err != nil {
		return nil, models.StoreSalesSummary{}, err
	}

	if err = tx.Commit(); err != nil {
		return nil, models.StoreSalesSummary{}, err
	}

	status, err = r.find(ctx, r.DB, branchID, businessDate, false)
	if err != nil {
		return nil, models.StoreSalesSummary{}, err
	}

	return status, summary, nil
}

type storeStatusQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func (r *StoreStatusRepository) find(
	ctx context.Context,
	queryer storeStatusQueryer,
	branchID int64,
	businessDate string,
	forUpdate bool,
) (*models.StoreStatus, error) {
	lockClause := ""
	if forUpdate {
		lockClause = " FOR UPDATE"
	}

	query := `
		SELECT
			ss.id,
			ss.business_date::TEXT,
			ss.store_id,
			ss.branch_id,
			ss.is_open,
			ss.cash_open,
			ss.cash_close,
			ss.expected_cash_at_close,
			ss.cash_difference,
			TO_CHAR(ss.scheduled_close_time, 'HH24:MI'),
			COALESCE(ss.closed_by, ss.opened_by) AS user_id,
			ss.opened_by,
			ss.closed_by,
			ss.opened_at,
			CASE
				WHEN ss.closed_at IS NOT NULL THEN TO_CHAR(ss.closed_at AT TIME ZONE 'Asia/Jakarta', 'HH24:MI:SS')
				ELSE TO_CHAR(ss.scheduled_close_time, 'HH24:MI:SS')
			END AS closed_at,
			ss.closed_at,
			ss.created_at,
			ss.updated_at
		FROM store_statuses ss
		WHERE ss.branch_id = $1
		AND ss.business_date = $2::DATE` + lockClause

	var status models.StoreStatus
	var cashOpen sql.NullFloat64
	var cashClose sql.NullFloat64
	var expectedCash sql.NullFloat64
	var cashDifference sql.NullFloat64
	var scheduledCloseTime sql.NullString
	var userID sql.NullInt64
	var openedBy sql.NullInt64
	var closedBy sql.NullInt64
	var openedAt sql.NullTime
	var closedAt sql.NullString
	var closedAtTimestamp sql.NullTime

	err := queryer.QueryRowContext(ctx, query, branchID, businessDate).Scan(
		&status.ID,
		&status.Date,
		&status.StoreID,
		&status.BranchID,
		&status.IsOpen,
		&cashOpen,
		&cashClose,
		&expectedCash,
		&cashDifference,
		&scheduledCloseTime,
		&userID,
		&openedBy,
		&closedBy,
		&openedAt,
		&closedAt,
		&closedAtTimestamp,
		&status.CreatedAt,
		&status.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	status.CashOpen = nullableFloat64(cashOpen)
	status.CashClose = nullableFloat64(cashClose)
	status.ExpectedCashAtClose = nullableFloat64(expectedCash)
	status.CashDifference = nullableFloat64(cashDifference)
	status.ScheduledCloseTime = helpers.NullableString(scheduledCloseTime)
	status.UserID = helpers.NullableInt64(userID)
	status.OpenedBy = helpers.NullableInt64(openedBy)
	status.ClosedBy = helpers.NullableInt64(closedBy)
	status.OpenedAt = helpers.NullableTime(openedAt)
	status.ClosedAt = helpers.NullableString(closedAt)
	status.ClosedAtTimestamp = helpers.NullableTime(closedAtTimestamp)

	return &status, nil
}

func (r *StoreStatusRepository) salesSummary(
	ctx context.Context,
	queryer storeStatusQueryer,
	branchID int64,
	businessDate string,
) (models.StoreSalesSummary, error) {
	query := `
		SELECT
			COALESCE(SUM(t.grand_total), 0) AS pos_sales,
			COALESCE(SUM(
				GREATEST(
					t.grand_total - GREATEST(LEAST(t.transfer_amount, t.grand_total), 0),
					0
				)
			), 0) AS pos_cash_total,
			COALESCE(SUM(
				GREATEST(LEAST(t.transfer_amount, t.grand_total), 0)
			), 0) AS pos_transfer_total
		FROM transactions t
		WHERE t.branch_id = $1
		AND DATE(t.transaction_date AT TIME ZONE 'Asia/Jakarta') = $2::DATE
		AND t.status = 'paid'
		AND t.deleted_at IS NULL
	`

	var summary models.StoreSalesSummary
	err := queryer.QueryRowContext(ctx, query, branchID, businessDate).Scan(
		&summary.POSSales,
		&summary.POSCashTotal,
		&summary.POSTransferTotal,
	)
	if err != nil {
		return models.StoreSalesSummary{}, err
	}

	summary.ExpectedCash = summary.POSCashTotal
	return summary, nil
}

func nullableFloat64(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	return &value.Float64
}

func nullableStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
