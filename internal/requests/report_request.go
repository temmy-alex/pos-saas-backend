package requests

type DailySalesReportRequest struct {
	Date     string `form:"date" binding:"required"`
	StoreID  int64  `form:"store_id"`
	BranchID int64  `form:"branch_id"`
}
