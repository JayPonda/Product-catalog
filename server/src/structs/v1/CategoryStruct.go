package structsv1

import (
	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/google/uuid"
)

// RequestCategory is the payload for creating a category.
type RequestCategory struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}

// LinkCategoryRequest is the payload for linking or unlinking
// an existing category to/from a product.
type LinkCategoryRequest struct {
	CategoryID uuid.UUID `json:"category_id" validate:"required"`
}

// ListCategoriesQuery holds pagination parameters for listing categories.
type ListCategoriesQuery struct {
	Limit  int `query:"limit" validate:"omitempty,oneof=20 50 100"`
	Offset int `query:"offset" validate:"omitempty,min=0"`
}

// ListCategoriesResponse is the paginated category list payload.
type ListCategoriesResponse struct {
	Categories []models.Category `json:"categories"`
	Total      int64             `json:"total"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
}

// MatchCategoriesQuery holds the prefix and limit for category matching.
type MatchCategoriesQuery struct {
	Name  string `query:"name" validate:"required,min=1,max=50"`
	Limit int    `query:"limit" validate:"omitempty,oneof=5 10 20"`
}

// MatchCategoriesResponse is the payload returned by the category match endpoint.
type MatchCategoriesResponse struct {
	Categories []models.Category `json:"categories"`
}
