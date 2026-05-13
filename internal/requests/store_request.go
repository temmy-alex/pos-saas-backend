package requests

type StoreRequest struct {
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	IsActive *bool  `json:"is_active"`
}
