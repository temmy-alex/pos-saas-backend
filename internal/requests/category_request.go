package requests

type CategoryRequest struct {
	StoreID     int64  `json:"store_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}
