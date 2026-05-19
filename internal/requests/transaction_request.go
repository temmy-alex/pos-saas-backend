package requests

type TransactionRequest struct {
	StoreID  int64 `json:"store_id" binding:"required"`
	BranchID int64 `json:"branch_id" binding:"required"`

	CustomerName  string `json:"customer_name"`
	PaymentMethod string `json:"payment_method" binding:"required"`

	CashAmount     float64 `json:"cash_amount"`
	TransferAmount float64 `json:"transfer_amount"`

	Notes string `json:"notes"`

	Items []TransactionItemRequest `json:"items" binding:"required,min=1,dive"`
}

type TransactionItemRequest struct {
	ProductID int64   `json:"product_id" binding:"required"`
	Qty       int     `json:"qty" binding:"required,gte=1"`
	Discount  float64 `json:"discount"`
}
