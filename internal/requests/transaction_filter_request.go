package requests

type TransactionFilterRequest struct {
	StoreID       int64  `form:"store_id"`
	BranchID      int64  `form:"branch_id"`
	CashierID     int64  `form:"cashier_id"`
	Date          string `form:"date"`
	PaymentMethod string `form:"payment_method"`
	Search        string `form:"search"`
	Page          int    `form:"page"`
	Limit         int    `form:"limit"`
}
