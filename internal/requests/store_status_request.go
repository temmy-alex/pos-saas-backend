package requests

type StoreStatusFilterRequest struct {
	BranchID int64  `form:"branch_id"`
	Date     string `form:"date"`
}

type OpenStoreRequest struct {
	CashOpen *float64 `json:"cash_open" binding:"required"`
	ClosedAt *string  `json:"closed_at"`
	BranchID int64    `json:"branch_id"`
}

type CloseStoreRequest struct {
	CashClose *float64 `json:"cash_close" binding:"required"`
	BranchID  int64    `json:"branch_id"`
}
