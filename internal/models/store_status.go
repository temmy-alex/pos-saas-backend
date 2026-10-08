package models

import "time"

type StoreStatus struct {
	ID                  int64      `json:"id"`
	Date                string     `json:"date"`
	StoreID             int64      `json:"store_id"`
	BranchID            int64      `json:"branch_id"`
	IsOpen              bool       `json:"is_open"`
	CashOpen            *float64   `json:"cash_open"`
	CashClose           *float64   `json:"cash_close"`
	ExpectedCashAtClose *float64   `json:"expected_cash_at_close"`
	CashDifference      *float64   `json:"cash_difference"`
	ScheduledCloseTime  *string    `json:"scheduled_close_time"`
	UserID              *int64     `json:"user_id"`
	OpenedBy            *int64     `json:"opened_by"`
	ClosedBy            *int64     `json:"closed_by"`
	OpenedAt            *time.Time `json:"opened_at"`
	ClosedAt            *string    `json:"closed_at"`
	ClosedAtTimestamp   *time.Time `json:"closed_at_timestamp"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type StoreSalesSummary struct {
	POSSales         float64 `json:"posSales"`
	POSCashTotal     float64 `json:"posCashTotal"`
	POSTransferTotal float64 `json:"posTransferTotal"`
	ExpectedCash     float64 `json:"expectedCash"`
}
