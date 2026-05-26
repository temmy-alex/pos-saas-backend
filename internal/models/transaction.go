package models

import "time"

type Transaction struct {
	ID int64 `json:"id"`

	StoreID   int64  `json:"store_id"`
	StoreName string `json:"store_name"`

	BranchID   int64  `json:"branch_id"`
	BranchName string `json:"branch_name"`

	CashierID   int64  `json:"cashier_id"`
	CashierName string `json:"cashier_name"`

	TransactionNumber string  `json:"transaction_number"`
	CustomerName      *string `json:"customer_name"`

	PaymentMethod string `json:"payment_method"`

	Subtotal      float64 `json:"subtotal"`
	DiscountTotal float64 `json:"discount_total"`
	GrandTotal    float64 `json:"grand_total"`

	CashAmount     float64 `json:"cash_amount"`
	TransferAmount float64 `json:"transfer_amount"`
	ChangeAmount   float64 `json:"change_amount"`

	Notes  *string `json:"notes"`
	Status string  `json:"status"`

	VoidReason   *string    `json:"void_reason"`
	VoidedAt     *time.Time `json:"voided_at"`
	VoidedBy     *int64     `json:"voided_by"`
	VoidedByName *string    `json:"voided_by_name"`

	TransactionDate time.Time `json:"transaction_date"`

	Items []TransactionItem `json:"items,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TransactionItem struct {
	ID int64 `json:"id"`

	TransactionID int64 `json:"transaction_id"`
	ProductID     int64 `json:"product_id"`

	ProductNameSnapshot string `json:"product_name_snapshot"`
	ProductSKUSnapshot  string `json:"product_sku_snapshot"`

	Qty      int     `json:"qty"`
	Price    float64 `json:"price"`
	Discount float64 `json:"discount"`
	Subtotal float64 `json:"subtotal"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
