package requests

type StockOpnameFilterRequest struct {
	StoreID  int64  `form:"store_id"`
	BranchID int64  `form:"branch_id"`
	Status   string `form:"status"`
	Search   string `form:"search"`
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	PerPage  int    `form:"per_page"`
}

type StockOpnameRequest struct {
	BranchID    int64                    `json:"branch_id"`
	Note        string                   `json:"note" binding:"max=255"`
	Items       []StockOpnameItemRequest `json:"items"`
	LegacyItems []StockOpnameItemRequest `json:"stockOpnameInput"`
}

func (r StockOpnameRequest) NormalizedItems() []StockOpnameItemRequest {
	if len(r.Items) > 0 {
		return r.Items
	}
	return r.LegacyItems
}

type StockOpnameItemRequest struct {
	ProductID int64  `json:"product_id" binding:"required"`
	OldStock  *int   `json:"old_stock"`
	NewStock  int    `json:"new_stock" binding:"gte=0"`
	Note      string `json:"note" binding:"max=255"`
}

type CancelStockOpnameRequest struct {
	Reason string `json:"reason" binding:"required,max=255"`
}
