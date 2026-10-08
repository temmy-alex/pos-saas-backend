package repositories

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/models"
	"pos-saas-backend/internal/requests"
)

var (
	ErrStockOpnameNotFound         = errors.New("stock opname not found")
	ErrStockOpnameAlreadyCompleted = errors.New("stock opname is already completed")
	ErrStockOpnameAlreadyCancelled = errors.New("stock opname is already cancelled")
)

type StockChangedError struct {
	ProductID int64
	Expected  int
	Actual    int
}

func (e *StockChangedError) Error() string {
	return fmt.Sprintf(
		"stock for product %d changed from %d to %d; create a new stock opname with the latest stock",
		e.ProductID,
		e.Expected,
		e.Actual,
	)
}

type StockOpnameRepository struct {
	DB *sql.DB
}

func NewStockOpnameRepository(db *sql.DB) *StockOpnameRepository {
	return &StockOpnameRepository{DB: db}
}

func (r *StockOpnameRepository) FindAll(
	ctx context.Context,
	filter requests.StockOpnameFilterRequest,
) ([]models.StockOpname, int64, error) {
	whereSQL, args, nextPosition := buildStockOpnameWhere(filter, "so", "")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM stock_opnames so WHERE %s", whereSQL)
	if err := r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	queryArgs := append(args, filter.Limit, offset)
	query := fmt.Sprintf(`
		SELECT
			so.id,
			so.opname_number,
			so.store_id,
			s.name AS store_name,
			so.branch_id,
			b.name AS branch_name,
			so.counted_by,
			cu.name AS counted_by_name,
			so.completed_by,
			completed_user.name AS completed_by_name,
			so.cancelled_by,
			cancelled_user.name AS cancelled_by_name,
			so.status,
			so.note,
			so.cancel_reason,
			(SELECT COUNT(*) FROM stock_opname_details d WHERE d.stock_opname_id = so.id) AS item_count,
			COALESCE((SELECT SUM(d.difference) FROM stock_opname_details d WHERE d.stock_opname_id = so.id), 0) AS total_difference,
			so.completed_at,
			so.cancelled_at,
			so.created_at,
			so.updated_at
		FROM stock_opnames so
		INNER JOIN stores s ON s.id = so.store_id
		INNER JOIN branches b ON b.id = so.branch_id
		INNER JOIN users cu ON cu.id = so.counted_by
		LEFT JOIN users completed_user ON completed_user.id = so.completed_by
		LEFT JOIN users cancelled_user ON cancelled_user.id = so.cancelled_by
		WHERE %s
		ORDER BY so.id DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, nextPosition, nextPosition+1)

	rows, err := r.DB.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	opnames := make([]models.StockOpname, 0)
	for rows.Next() {
		opname, err := scanStockOpname(rows)
		if err != nil {
			return nil, 0, err
		}
		opnames = append(opnames, *opname)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return opnames, total, nil
}

func (r *StockOpnameRepository) FindMobileDetails(
	ctx context.Context,
	filter requests.StockOpnameFilterRequest,
) ([]models.StockOpnameMobileDetail, int64, error) {
	whereSQL, args, nextPosition := buildStockOpnameWhere(filter, "so", "d")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM stock_opname_details d
		INNER JOIN stock_opnames so ON so.id = d.stock_opname_id
		WHERE %s
	`, whereSQL)
	var total int64
	if err := r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	queryArgs := append(args, filter.Limit, offset)
	query := fmt.Sprintf(`
		SELECT
			d.id,
			d.stock_opname_id,
			d.product_id,
			d.store_id,
			d.branch_id,
			d.old_stock,
			d.new_stock,
			d.difference,
			p.id,
			p.name,
			so.id,
			so.opname_number,
			so.status,
			b.id,
			b.name,
			b.store_id,
			u.id,
			u.name,
			d.created_at,
			d.updated_at
		FROM stock_opname_details d
		INNER JOIN stock_opnames so ON so.id = d.stock_opname_id
		INNER JOIN products p ON p.id = d.product_id
		INNER JOIN branches b ON b.id = so.branch_id
		INNER JOIN users u ON u.id = so.counted_by
		WHERE %s
		ORDER BY d.id DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, nextPosition, nextPosition+1)

	rows, err := r.DB.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	details := make([]models.StockOpnameMobileDetail, 0)
	for rows.Next() {
		var detail models.StockOpnameMobileDetail
		err := rows.Scan(
			&detail.ID,
			&detail.StockOpnameID,
			&detail.ProductID,
			&detail.StoreID,
			&detail.BranchID,
			&detail.OldStock,
			&detail.NewStock,
			&detail.Difference,
			&detail.Product.ID,
			&detail.Product.Name,
			&detail.StockOpname.ID,
			&detail.StockOpname.OpnameNumber,
			&detail.StockOpname.Status,
			&detail.StockOpname.Branch.ID,
			&detail.StockOpname.Branch.Name,
			&detail.StockOpname.Branch.StoreID,
			&detail.StockOpname.User.ID,
			&detail.StockOpname.User.Name,
			&detail.CreatedAt,
			&detail.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		details = append(details, detail)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return details, total, nil
}

func (r *StockOpnameRepository) FindByID(ctx context.Context, id int64) (*models.StockOpname, error) {
	query := `
		SELECT
			so.id,
			so.opname_number,
			so.store_id,
			s.name AS store_name,
			so.branch_id,
			b.name AS branch_name,
			so.counted_by,
			cu.name AS counted_by_name,
			so.completed_by,
			completed_user.name AS completed_by_name,
			so.cancelled_by,
			cancelled_user.name AS cancelled_by_name,
			so.status,
			so.note,
			so.cancel_reason,
			(SELECT COUNT(*) FROM stock_opname_details d WHERE d.stock_opname_id = so.id) AS item_count,
			COALESCE((SELECT SUM(d.difference) FROM stock_opname_details d WHERE d.stock_opname_id = so.id), 0) AS total_difference,
			so.completed_at,
			so.cancelled_at,
			so.created_at,
			so.updated_at
		FROM stock_opnames so
		INNER JOIN stores s ON s.id = so.store_id
		INNER JOIN branches b ON b.id = so.branch_id
		INNER JOIN users cu ON cu.id = so.counted_by
		LEFT JOIN users completed_user ON completed_user.id = so.completed_by
		LEFT JOIN users cancelled_user ON cancelled_user.id = so.cancelled_by
		WHERE so.id = $1
	`

	opname, err := scanStockOpname(r.DB.QueryRowContext(ctx, query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	items, err := r.findItems(ctx, id)
	if err != nil {
		return nil, err
	}
	opname.Items = items

	return opname, nil
}

func (r *StockOpnameRepository) Create(
	ctx context.Context,
	storeID, branchID, userID int64,
	request requests.StockOpnameRequest,
) (*models.StockOpname, error) {
	items := append([]requests.StockOpnameItemRequest(nil), request.NormalizedItems()...)
	if len(items) == 0 {
		return nil, errors.New("at least one stock opname item is required")
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ProductID < items[j].ProductID
	})

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	opnameNumber, err := generateStockOpnameNumber()
	if err != nil {
		return nil, err
	}

	var opnameID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO stock_opnames (
			opname_number, store_id, branch_id, counted_by, status, note
		) VALUES ($1, $2, $3, $4, 'draft', NULLIF($5, ''))
		RETURNING id
	`, opnameNumber, storeID, branchID, userID, strings.TrimSpace(request.Note)).Scan(&opnameID)
	if err != nil {
		return nil, err
	}

	seenProducts := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if _, exists := seenProducts[item.ProductID]; exists {
			return nil, fmt.Errorf("product_id %d is duplicated", item.ProductID)
		}
		seenProducts[item.ProductID] = struct{}{}

		var currentStock int
		err = tx.QueryRowContext(ctx, `
			SELECT stock
			FROM products
			WHERE id = $1
			AND store_id = $2
			AND branch_id = $3
			AND deleted_at IS NULL
			FOR UPDATE
		`, item.ProductID, storeID, branchID).Scan(&currentStock)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("product %d not found in the selected branch", item.ProductID)
			}
			return nil, err
		}

		if item.OldStock != nil && *item.OldStock != currentStock {
			return nil, &StockChangedError{
				ProductID: item.ProductID,
				Expected:  *item.OldStock,
				Actual:    currentStock,
			}
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO stock_opname_details (
				stock_opname_id, product_id, store_id, branch_id,
				old_stock, new_stock, difference, note
			) VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''))
		`,
			opnameID,
			item.ProductID,
			storeID,
			branchID,
			currentStock,
			item.NewStock,
			item.NewStock-currentStock,
			strings.TrimSpace(item.Note),
		)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.FindByID(ctx, opnameID)
}

func (r *StockOpnameRepository) Complete(ctx context.Context, id, userID int64) (*models.StockOpname, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var status string
	err = tx.QueryRowContext(ctx, `
		SELECT status
		FROM stock_opnames
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrStockOpnameNotFound
		}
		return nil, err
	}
	if status == "completed" {
		return nil, ErrStockOpnameAlreadyCompleted
	}
	if status == "cancelled" {
		return nil, ErrStockOpnameAlreadyCancelled
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT product_id, store_id, branch_id, old_stock, new_stock
		FROM stock_opname_details
		WHERE stock_opname_id = $1
		ORDER BY product_id
	`, id)
	if err != nil {
		return nil, err
	}

	type completionItem struct {
		ProductID int64
		StoreID   int64
		BranchID  int64
		OldStock  int
		NewStock  int
	}
	items := make([]completionItem, 0)
	for rows.Next() {
		var item completionItem
		if err = rows.Scan(&item.ProductID, &item.StoreID, &item.BranchID, &item.OldStock, &item.NewStock); err != nil {
			_ = rows.Close()
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("stock opname has no items")
	}

	for _, item := range items {
		var currentStock int
		err = tx.QueryRowContext(ctx, `
			SELECT stock
			FROM products
			WHERE id = $1
			AND store_id = $2
			AND branch_id = $3
			AND deleted_at IS NULL
			FOR UPDATE
		`, item.ProductID, item.StoreID, item.BranchID).Scan(&currentStock)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("product %d not found in the selected branch", item.ProductID)
			}
			return nil, err
		}
		if currentStock != item.OldStock {
			return nil, &StockChangedError{
				ProductID: item.ProductID,
				Expected:  item.OldStock,
				Actual:    currentStock,
			}
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE products
			SET stock = $1, updated_at = NOW()
			WHERE id = $2
		`, item.NewStock, item.ProductID)
		if err != nil {
			return nil, err
		}
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE stock_opnames
		SET
			status = 'completed',
			completed_by = $1,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE id = $2
	`, userID, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *StockOpnameRepository) Cancel(
	ctx context.Context,
	id, userID int64,
	reason string,
) (*models.StockOpname, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var status string
	err = tx.QueryRowContext(ctx, `
		SELECT status
		FROM stock_opnames
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrStockOpnameNotFound
		}
		return nil, err
	}
	if status == "completed" {
		return nil, ErrStockOpnameAlreadyCompleted
	}
	if status == "cancelled" {
		return nil, ErrStockOpnameAlreadyCancelled
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE stock_opnames
		SET
			status = 'cancelled',
			cancelled_by = $1,
			cancel_reason = $2,
			cancelled_at = NOW(),
			updated_at = NOW()
		WHERE id = $3
	`, userID, strings.TrimSpace(reason), id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *StockOpnameRepository) findItems(ctx context.Context, stockOpnameID int64) ([]models.StockOpnameItem, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT
			d.id,
			d.stock_opname_id,
			d.product_id,
			p.name,
			p.sku,
			d.store_id,
			d.branch_id,
			d.old_stock,
			d.new_stock,
			d.difference,
			d.note,
			d.created_at,
			d.updated_at
		FROM stock_opname_details d
		INNER JOIN products p ON p.id = d.product_id
		WHERE d.stock_opname_id = $1
		ORDER BY d.id
	`, stockOpnameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.StockOpnameItem, 0)
	for rows.Next() {
		var item models.StockOpnameItem
		var note sql.NullString
		err := rows.Scan(
			&item.ID,
			&item.StockOpnameID,
			&item.ProductID,
			&item.ProductName,
			&item.ProductSKU,
			&item.StoreID,
			&item.BranchID,
			&item.OldStock,
			&item.NewStock,
			&item.Difference,
			&note,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		item.Note = helpers.NullableString(note)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func buildStockOpnameWhere(
	filter requests.StockOpnameFilterRequest,
	alias string,
	detailAlias string,
) (string, []interface{}, int) {
	clauses := []string{"1 = 1"}
	args := make([]interface{}, 0)
	position := 1

	if filter.StoreID > 0 {
		clauses = append(clauses, fmt.Sprintf("%s.store_id = $%d", alias, position))
		args = append(args, filter.StoreID)
		position++
	}
	if filter.BranchID > 0 {
		clauses = append(clauses, fmt.Sprintf("%s.branch_id = $%d", alias, position))
		args = append(args, filter.BranchID)
		position++
	}
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("%s.status = $%d", alias, position))
		args = append(args, filter.Status)
		position++
	}
	if filter.Search != "" {
		if detailAlias == "" {
			clauses = append(clauses, fmt.Sprintf(
				"(%s.opname_number ILIKE $%d OR %s.note ILIKE $%d)",
				alias,
				position,
				alias,
				position,
			))
		} else {
			clauses = append(clauses, fmt.Sprintf(
				"(%s.opname_number ILIKE $%d OR %s.note ILIKE $%d OR %s.old_stock::TEXT ILIKE $%d OR %s.new_stock::TEXT ILIKE $%d)",
				alias,
				position,
				alias,
				position,
				detailAlias,
				position,
				detailAlias,
				position,
			))
		}
		args = append(args, "%"+filter.Search+"%")
		position++
	}

	return strings.Join(clauses, " AND "), args, position
}

type stockOpnameScanner interface {
	Scan(dest ...interface{}) error
}

func scanStockOpname(scanner stockOpnameScanner) (*models.StockOpname, error) {
	var opname models.StockOpname
	var completedBy sql.NullInt64
	var completedByName sql.NullString
	var cancelledBy sql.NullInt64
	var cancelledByName sql.NullString
	var note sql.NullString
	var cancelReason sql.NullString
	var completedAt sql.NullTime
	var cancelledAt sql.NullTime

	err := scanner.Scan(
		&opname.ID,
		&opname.OpnameNumber,
		&opname.StoreID,
		&opname.StoreName,
		&opname.BranchID,
		&opname.BranchName,
		&opname.CountedBy,
		&opname.CountedByName,
		&completedBy,
		&completedByName,
		&cancelledBy,
		&cancelledByName,
		&opname.Status,
		&note,
		&cancelReason,
		&opname.ItemCount,
		&opname.TotalDifference,
		&completedAt,
		&cancelledAt,
		&opname.CreatedAt,
		&opname.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	opname.CompletedBy = helpers.NullableInt64(completedBy)
	opname.CompletedByName = helpers.NullableString(completedByName)
	opname.CancelledBy = helpers.NullableInt64(cancelledBy)
	opname.CancelledByName = helpers.NullableString(cancelledByName)
	opname.Note = helpers.NullableString(note)
	opname.CancelReason = helpers.NullableString(cancelReason)
	opname.CompletedAt = helpers.NullableTime(completedAt)
	opname.CancelledAt = helpers.NullableTime(cancelledAt)

	return &opname, nil
}

func generateStockOpnameNumber() (string, error) {
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"SO-%s-%s",
		time.Now().Format("20060102"),
		strings.ToUpper(hex.EncodeToString(randomBytes)),
	), nil
}
