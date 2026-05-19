package models

type DashboardSummary struct {
	Date              string  `json:"date"`
	StoreID           int64   `json:"store_id"`
	BranchID          int64   `json:"branch_id"`
	TotalTransactions int64   `json:"total_transactions"`
	TotalItemsSold    int64   `json:"total_items_sold"`
	TotalSales        float64 `json:"total_sales"`
	TotalCash         float64 `json:"total_cash"`
	TotalTransfer     float64 `json:"total_transfer"`
	TotalDiscount     float64 `json:"total_discount"`
	LowStockCount     int64   `json:"low_stock_count"`
}

type DashboardHourlySales struct {
	Hour              int     `json:"hour"`
	Label             string  `json:"label"`
	TotalTransactions int64   `json:"total_transactions"`
	TotalSales        float64 `json:"total_sales"`
}

type DashboardResponse struct {
	Summary            DashboardSummary       `json:"summary"`
	SalesPerHour       []DashboardHourlySales `json:"sales_per_hour"`
	LowStockProducts   []Product              `json:"low_stock_products"`
	RecentTransactions []Transaction          `json:"recent_transactions"`
}
