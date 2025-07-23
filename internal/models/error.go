package models

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrConflictData           = errors.New("data conflicts with existing data")
	ErrDataNotFound           = errors.New("data not found")
	ErrInvalidCredentials     = errors.New("invalid login or password")
	ErrInvalidOrderNumber     = errors.New("invalid order number")
	ErrOrderLoadedUser        = errors.New("order already loaded by user")
	ErrOrderLoadedAnotherUser = errors.New("order already loaded by another user")
	ErrOrderNotRegInAccrual   = errors.New("order is not registered in the accrual")
	ErrInternalError          = errors.New("internal error")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrOrderExist             = errors.New("order already exists")
	ErrWithdrawalsNotExist    = errors.New("withdrawals not exists")
)

type TooManyRequestsError struct {
	RetryAfter time.Duration
}

func (e TooManyRequestsError) Error() string {
	return fmt.Sprintf("Error: too many requests, retry after %v", e.RetryAfter)
}

func NewTooManyRequestsError(retryAfter time.Duration) *TooManyRequestsError {
	return &TooManyRequestsError{RetryAfter: retryAfter}
}
