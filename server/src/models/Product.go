package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	Name          string        `json:"name" db:"name"`
	Description   string        `json:"description" db:"description"`
	Price         int64         `json:"price" db:"price"`
	StockQuantity int           `json:"stock_quantity" db:"stock_quantity"`
	UserID        uuid.NullUUID `json:"user_id" db:"user_id"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
	DeletedAt     sql.NullTime  `json:"deleted_at" db:"deleted_at"`

	// category fields
	Categories []Category `json:"categories" db:"-"`
}
