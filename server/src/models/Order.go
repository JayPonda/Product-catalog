package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Order represents a customer order. It supports soft deletion via deleted_at,
// consistent with the other tables in this project.
type Order struct {
	ID         uuid.UUID    `json:"id" db:"id"`
	CustomerID uuid.UUID    `json:"customer_id" db:"customer_id"`
	TotalBill  float64      `json:"total_bill" db:"total_bill"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt  sql.NullTime `json:"deleted_at" db:"deleted_at"`
}
