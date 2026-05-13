package requests

type BranchRequest struct {
	StoreID  int64  `json:"store_id" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	IsActive *bool  `json:"is_active"`
}
