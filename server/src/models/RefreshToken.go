package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	UserID    uuid.UUID    `json:"user_id" db:"user_id"`
	TokenHash string       `json:"-" db:"token_hash"`
	ExpiresAt time.Time    `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	DeletedAt sql.NullTime `json:"-" db:"deleted_at"`
}
