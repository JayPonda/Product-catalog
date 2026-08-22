package structsv1

import (
	"github.com/JayPonda/Product-catalog/server/src/models"
)

// ListOrdersQuery holds pagination parameters for listing orders.
type ListOrdersQuery struct {
	Limit  int `query:"limit" validate:"omitempty,oneof=20 50 100"`
	Offset int `query:"offset" validate:"omitempty,min=0"`
}

// ListOrdersResponse is the paginated order list payload.
type ListOrdersResponse struct {
	Orders []models.Order `json:"orders"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}
