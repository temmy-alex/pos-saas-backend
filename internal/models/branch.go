package models

import "time"

type Branch struct {
	ID        int64     `json:"id"`
	StoreID   int64     `json:"store_id"`
	StoreName string    `json:"store_name"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Address   *string   `json:"address"`
	Phone     *string   `json:"phone"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
