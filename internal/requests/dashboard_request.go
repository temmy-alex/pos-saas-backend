package requests

type DashboardRequest struct {
	Date     string `form:"date"`
	StoreID  int64  `form:"store_id"`
	BranchID int64  `form:"branch_id"`
}
