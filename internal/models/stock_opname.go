package models

import "time"

type StockOpname struct {
	ID              int64             `json:"id"`
	OpnameNumber    string            `json:"opname_number"`
	StoreID         int64             `json:"store_id"`
	StoreName       string            `json:"store_name"`
	BranchID        int64             `json:"branch_id"`
	BranchName      string            `json:"branch_name"`
	CountedBy       int64             `json:"counted_by"`
	CountedByName   string            `json:"counted_by_name"`
	CompletedBy     *int64            `json:"completed_by"`
	CompletedByName *string           `json:"completed_by_name"`
	CancelledBy     *int64            `json:"cancelled_by"`
	CancelledByName *string           `json:"cancelled_by_name"`
	Status          string            `json:"status"`
	Note            *string           `json:"note"`
	CancelReason    *string           `json:"cancel_reason"`
	ItemCount       int64             `json:"item_count"`
	TotalDifference int64             `json:"total_difference"`
	CompletedAt     *time.Time        `json:"completed_at"`
	CancelledAt     *time.Time        `json:"cancelled_at"`
	Items           []StockOpnameItem `json:"items,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type StockOpnameItem struct {
	ID            int64     `json:"id"`
	StockOpnameID int64     `json:"stock_opname_id"`
	ProductID     int64     `json:"product_id"`
	ProductName   string    `json:"product_name"`
	ProductSKU    string    `json:"product_sku"`
	StoreID       int64     `json:"store_id"`
	BranchID      int64     `json:"branch_id"`
	OldStock      int       `json:"old_stock"`
	NewStock      int       `json:"new_stock"`
	Difference    int       `json:"difference"`
	Note          *string   `json:"note"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type StockOpnameMobileDetail struct {
	ID            int64 `json:"id"`
	StockOpnameID int64 `json:"stock_opname_id"`
	ProductID     int64 `json:"product_id"`
	StoreID       int64 `json:"store_id"`
	BranchID      int64 `json:"branch_id"`
	OldStock      int   `json:"old_stock"`
	NewStock      int   `json:"new_stock"`
	Difference    int   `json:"difference"`
	Product       struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"product"`
	StockOpname struct {
		ID           int64  `json:"id"`
		OpnameNumber string `json:"opname_number"`
		Status       string `json:"status"`
		Branch       struct {
			ID      int64  `json:"id"`
			Name    string `json:"name"`
			StoreID int64  `json:"store_id"`
		} `json:"branch"`
		User struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"user"`
	} `json:"stock_opname"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
