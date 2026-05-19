package models

type DailySalesReport struct {
	Date              string  `json:"date"`
	StoreID           int64   `json:"store_id"`
	BranchID          int64   `json:"branch_id"`
	TotalTransactions int64   `json:"total_transactions"`
	TotalItemsSold    int64   `json:"total_items_sold"`
	Subtotal          float64 `json:"subtotal"`
	DiscountTotal     float64 `json:"discount_total"`
	GrandTotal        float64 `json:"grand_total"`
	CashTotal         float64 `json:"cash_total"`
	TransferTotal     float64 `json:"transfer_total"`
	ChangeTotal       float64 `json:"change_total"`
}

type DailySalesPaymentBreakdown struct {
	PaymentMethod     string  `json:"payment_method"`
	TotalTransactions int64   `json:"total_transactions"`
	GrandTotal        float64 `json:"grand_total"`
	CashTotal         float64 `json:"cash_total"`
	TransferTotal     float64 `json:"transfer_total"`
}

type DailySalesTopProduct struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductSKU  string  `json:"product_sku"`
	TotalQty    int64   `json:"total_qty"`
	TotalSales  float64 `json:"total_sales"`
}
