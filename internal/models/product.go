package models

import "time"

type Product struct {
	ID int64 `json:"id"`

	StoreID   int64  `json:"store_id"`
	StoreName string `json:"store_name"`

	BranchID   int64  `json:"branch_id"`
	BranchName string `json:"branch_name"`

	CategoryID   int64  `json:"category_id"`
	CategoryName string `json:"category_name"`

	Name        string  `json:"name"`
	SKU         string  `json:"sku"`
	Barcode     *string `json:"barcode"`
	Description *string `json:"description"`

	CostPrice    float64 `json:"cost_price"`
	SellingPrice float64 `json:"selling_price"`

	Stock    int `json:"stock"`
	MinStock int `json:"min_stock"`

	ImagePath *string `json:"image_path"`
	ImageURL  *string `json:"image_url"`
	ImageDisk *string `json:"image_disk"`

	IsActive bool `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
