package repositories

import (
	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/google/uuid"
)

type CategoryRepository interface {
	GetCategoryyId(id uuid.UUID) *models.Category
	GetCategoryByName(name string) (bool, error)
	CreateCategory(product models.Category) (uuid.UUID, error)
	DeleteCategory(id uuid.UUID) error
}
