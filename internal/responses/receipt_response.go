package responses

import "time"

type ReceiptResponse struct {
	Store       ReceiptStore       `json:"store"`
	Branch      ReceiptBranch      `json:"branch"`
	Cashier     ReceiptCashier     `json:"cashier"`
	Transaction ReceiptTransaction `json:"transaction"`
	Items       []ReceiptItem      `json:"items"`
	Summary     ReceiptSummary     `json:"summary"`
	Payment     ReceiptPayment     `json:"payment"`
	Void        *ReceiptVoid       `json:"void"`
}

type ReceiptStore struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ReceiptBranch struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ReceiptCashier struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ReceiptTransaction struct {
	ID                int64     `json:"id"`
	TransactionNumber string    `json:"transaction_number"`
	TransactionDate   time.Time `json:"transaction_date"`
	CustomerName      *string   `json:"customer_name"`
	Status            string    `json:"status"`
	Notes             *string   `json:"notes"`
}

type ReceiptItem struct {
	ID          int64   `json:"id"`
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductSKU  string  `json:"product_sku"`
	Qty         int     `json:"qty"`
	Price       float64 `json:"price"`
	GrossTotal  float64 `json:"gross_total"`
	Discount    float64 `json:"discount"`
	Subtotal    float64 `json:"subtotal"`
}

type ReceiptSummary struct {
	Subtotal      float64 `json:"subtotal"`
	DiscountTotal float64 `json:"discount_total"`
	GrandTotal    float64 `json:"grand_total"`
}

type ReceiptPayment struct {
	PaymentMethod  string  `json:"payment_method"`
	CashAmount     float64 `json:"cash_amount"`
	TransferAmount float64 `json:"transfer_amount"`
	PaidAmount     float64 `json:"paid_amount"`
	ChangeAmount   float64 `json:"change_amount"`
}

type ReceiptVoid struct {
	VoidReason   *string    `json:"void_reason"`
	VoidedAt     *time.Time `json:"voided_at"`
	VoidedBy     *int64     `json:"voided_by"`
	VoidedByName *string    `json:"voided_by_name"`
}
