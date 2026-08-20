package v1

import "github.com/JayPonda/Product-catalog/server/src/models"

type Categories struct {
	New []string          `json:"new_categories"`
	Old []models.Category `json:"old_categories"`
}
