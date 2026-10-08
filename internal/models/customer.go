package models

import "time"

type Customer struct {
	ID         int64     `json:"id"`
	StoreID    int64     `json:"store_id"`
	StoreName  string    `json:"store_name"`
	BranchID   int64     `json:"branch_id"`
	BranchName string    `json:"branch_name"`
	Name       string    `json:"name"`
	Phone      string    `json:"phone"`
	Email      *string   `json:"email"`
	VisitCount int64     `json:"visit_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
