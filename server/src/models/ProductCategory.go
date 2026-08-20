package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ProductCategory struct {
	ID         uuid.UUID    `json:"id" db:"id"`
	ProductID  uuid.UUID    `json:"product_id" db:"product_id"`
	CategoryID uuid.UUID    `json:"category_id" db:"category_id"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt  sql.NullTime `json:"deleted_at" db:"deleted_at"`
}
