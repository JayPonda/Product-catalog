package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	Name      string       `json:"name" db:"name"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" db:"deleted_at"`
}
