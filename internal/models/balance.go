package models

import (
	"time"
)

// Balance contains current balance and Withdrawn
type Balance struct {
	Current   float64
	Withdrawn float64
}

// Withdraw is entity withdraw
type Withdraw struct {
	ID          uint64
	UserID      uint64
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}
