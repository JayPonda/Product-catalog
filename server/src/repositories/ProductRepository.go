package repositories

import (
	"database/sql"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

type ProductRepository struct {
	Db     *goqu.Database
	Logger *utils.StructuredLogger
}

const PRODUCT_DB = "products"

func InitProductRepository(db *goqu.Database, logger *utils.StructuredLogger) (*ProductRepository, error) {

	return &ProductRepository{
		Db:     db,
		Logger: logger,
	}, nil
}

func (ProductRepositoryPtr *ProductRepository) GetProductById(id uuid.UUID) (models.Product, error) {
	var product models.Product

	found, err := ProductRepositoryPtr.Db.
		From(PRODUCT_DB).
		Where(
			goqu.C("id").Eq(id),
			goqu.C("deleted_at").IsNull(),
		).
		Select(
			"id",
			"name",
			"description",
			"price",
			"stock_quantity",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStruct(&product)

	if err != nil {
		return product, err
	}

	if !found {
		return product, sql.ErrNoRows
	}

	return product, nil
}

func (ProductRepositoryPtr *ProductRepository) GetProductByName(name string) (models.Product, error) {
	var product models.Product

	found, err := ProductRepositoryPtr.Db.
		From(PRODUCT_DB).
		Where(
			goqu.C("name").Eq(name),
			goqu.C("deleted_at").IsNull(),
		).
		Select(
			"id",
			"name",
			"description",
			"price",
			"stock_quantity",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStruct(&product)

	if err != nil {
		return product, err
	}

	if !found {
		return product, sql.ErrNoRows
	}

	return product, nil
}

func (ProductRepositoryPtr *ProductRepository) CreateProduct(product models.Product) (models.Product, error) {
	uuid, err := utils.GetUUID()
	if err != nil {
		return product, err
	}

	_, err = ProductRepositoryPtr.Db.Insert(PRODUCT_DB).Rows(
		goqu.Record{
			"id":             uuid,
			"name":           product.Name,
			"description":    product.Description,
			"price":          product.Price,
			"stock_quantity": product.StockQuantity,
		},
	).Executor().Exec()

	if err != nil {
		return product, err
	}

	product, err = ProductRepositoryPtr.GetProductById(uuid)
	return product, err
}

func (ProductRepositoryPtr *ProductRepository) UpdateProduct(id uuid.UUID, preUpdateProduct models.Product) (models.Product, error) {

	_, err := ProductRepositoryPtr.Db.Update(PRODUCT_DB).Set(
		goqu.Record{
			"name":           preUpdateProduct.Name,
			"description":    preUpdateProduct.Description,
			"price":          preUpdateProduct.Price,
			"stock_quantity": preUpdateProduct.StockQuantity,
			"updated_at":     time.Now(),
		},
	).Where(
		goqu.C("id").Eq(id),
		goqu.C("deleted_at").IsNull(),
	).Executor().Exec()

	if err != nil {
		return preUpdateProduct, err
	}

	product, err := ProductRepositoryPtr.GetProductById(id)
	return product, err
}

func (ProductRepositoryPtr *ProductRepository) DeleteProduct(id uuid.UUID) error {
	_, err := ProductRepositoryPtr.Db.Update(PRODUCT_DB).Set(
		goqu.Record{
			"deleted_at": time.Now(),
		},
	).Where(
		goqu.C("id").Eq(id),
		goqu.C("deleted_at").IsNull(),
	).Executor().Exec()

	return err
}
