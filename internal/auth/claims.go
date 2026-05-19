package auth

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserID   int64  `json:"user_id"`
	StoreID  *int64 `json:"store_id"`
	BranchID *int64 `json:"branch_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
