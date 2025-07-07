package models

import (
	"github.com/google/uuid"
	"time"
)

// User is user entity
type User struct {
	ID        uint64
	Login     string
	Password  string
	CreatedAt time.Time
}

// TokenPayload is payload contains user id's
type TokenPayload struct {
	ID     uuid.UUID
	UserID uint64
}
