package requests

type ProductFilterRequest struct {
	StoreID    int64  `form:"store_id"`
	BranchID   int64  `form:"branch_id"`
	CategoryID int64  `form:"category_id"`
	Search     string `form:"search"`
	IsActive   *bool  `form:"is_active"`
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
}
