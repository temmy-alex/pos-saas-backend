package responses

type AuthUserResponse struct {
	ID       int64  `json:"id"`
	StoreID  *int64 `json:"store_id"`
	BranchID *int64 `json:"branch_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type LoginResponse struct {
	AccessToken string           `json:"access_token"`
	TokenType   string           `json:"token_type"`
	ExpiresIn   int64            `json:"expires_in"`
	User        AuthUserResponse `json:"user"`
}
