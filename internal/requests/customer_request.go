package requests

type CustomerRequest struct {
	BranchID int64  `json:"branch_id"`
	Name     string `json:"name" binding:"required,max=150"`
	Phone    string `json:"phone" binding:"required,max=50"`
	Email    string `json:"email" binding:"omitempty,email,max=150"`
}

type CustomerFilterRequest struct {
	StoreID  int64  `form:"store_id"`
	BranchID int64  `form:"branch_id"`
	Search   string `form:"search"`
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	PerPage  int    `form:"per_page"`
}
