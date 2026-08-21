package structsv1

import "github.com/JayPonda/Product-catalog/server/src/models"

// RequestCategories holds the new category names to create and existing categories to link.
type RequestCategories struct {
	New []string          `json:"new_categories"`
	Old []models.Category `json:"old_categories"`
}

// MatchCategoriesQuery holds the prefix and limit for category matching.
type MatchCategoriesQuery struct {
	Name  string `query:"name" validate:"required,min=1,max=50"`
	Limit int    `query:"limit" validate:"omitempty,min=1,max=100"`
}

// MatchCategoriesResponse is the payload returned by the category match endpoint.
type MatchCategoriesResponse struct {
	Categories []models.Category `json:"categories"`
}
