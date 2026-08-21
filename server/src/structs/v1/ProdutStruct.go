package structsv1

import (
	"github.com/JayPonda/Product-catalog/server/src/models"
)

// ResponseProduct represents a product with its categories, used for read responses.
type ResponseProduct struct {
	models.Product
	Categories []models.Category `json:"categories"`
}

// ListProductsQuery holds pagination parameters for listing products.
type ListProductsQuery struct {
	Limit  int `query:"limit" validate:"omitempty,oneof=20 50 100"`
	Offset int `query:"offset" validate:"omitempty,min=0"`
}

// ListProductsResponse is the paginated product list payload.
type ListProductsResponse struct {
	Products []ResponseProduct `json:"products"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// RequestProduct is the payload for creating or updating a product.
// Categories are managed separately via link/unlink routes.
type RequestProduct struct {
	Name          string `json:"name" db:"name" validate:"required,min=1,max=50,letter,printable"`
	Description   string `json:"description" db:"description" validate:"required,max=150,letter,printable"`
	Price         int64  `json:"price" db:"price" validate:"required,gt=0,max=999999999"`
	StockQuantity int    `json:"stock_quantity" db:"stock_quantity" validate:"min=1,max=2147483647"`
}
