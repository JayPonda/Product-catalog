package repositories

import (
	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/google/uuid"
)

type ProductRepository interface {
	GetProductById(id uuid.UUID) *models.Product
	IsProductExists(name string) (bool, error)
	CreateProduct(product models.Product) (uuid.UUID, error)
	UpdateProduct(id uuid.UUID, fullNewObject *models.Product) (uuid.UUID, error)
	DeleteProduct(id uuid.UUID) error
}
