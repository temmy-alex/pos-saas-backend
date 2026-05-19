package requests

type ProductRequest struct {
	StoreID    int64 `json:"store_id" form:"store_id" binding:"required"`
	BranchID   int64 `json:"branch_id" form:"branch_id" binding:"required"`
	CategoryID int64 `json:"category_id" form:"category_id" binding:"required"`

	Name        string `json:"name" form:"name" binding:"required"`
	SKU         string `json:"sku" form:"sku" binding:"required"`
	Barcode     string `json:"barcode" form:"barcode"`
	Description string `json:"description" form:"description"`

	CostPrice    float64 `json:"cost_price" form:"cost_price"`
	SellingPrice float64 `json:"selling_price" form:"selling_price" binding:"required"`

	Stock    int `json:"stock" form:"stock"`
	MinStock int `json:"min_stock" form:"min_stock"`

	ImagePath *string `json:"-"`
	ImageURL  *string `json:"-"`
	ImageDisk *string `json:"-"`

	IsActive *bool `json:"is_active" form:"is_active"`
}
