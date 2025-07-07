package models

import "errors"

var (
	ErrConflictData       = errors.New("data conflicts with existing data")
	ErrDataNotFound       = errors.New("data not found")
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrInvalidOrderID     = errors.New("invalid order id")
)
