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

func (ProductRepositoryPtr *ProductRepository) GetProductById(id uuid.UUID, exec ...utils.Executor) (models.Product, error) {
	var product models.Product

	db := utils.ResolveExecutor(ProductRepositoryPtr.Db, exec)

	found, err := db.
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

func (ProductRepositoryPtr *ProductRepository) GetProductByName(name string, exec ...utils.Executor) (models.Product, error) {
	var product models.Product

	db := utils.ResolveExecutor(ProductRepositoryPtr.Db, exec)

	found, err := db.
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

func (ProductRepositoryPtr *ProductRepository) GetProducts(limit int, offset int, exec ...utils.Executor) ([]models.Product, int64, error) {
	db := utils.ResolveExecutor(ProductRepositoryPtr.Db, exec)

	var products []models.Product

	err := db.
		From(PRODUCT_DB).
		Where(
			goqu.C("deleted_at").IsNull(),
		).
		Order(goqu.I("created_at").Desc()).
		Limit(uint(limit)).
		Offset(uint(offset)).
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
		ScanStructs(&products)

	if err != nil {
		return nil, 0, err
	}

	var total int64

	_, err = db.
		From(PRODUCT_DB).
		Where(
			goqu.C("deleted_at").IsNull(),
		).
		Select(goqu.COUNT("*")).
		ScanVal(&total)

	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (ProductRepositoryPtr *ProductRepository) CreateProduct(product models.Product, exec ...utils.Executor) (models.Product, error) {
	uuid, err := utils.GetUUID()
	if err != nil {
		return product, err
	}

	db := utils.ResolveExecutor(ProductRepositoryPtr.Db, exec)

	_, err = db.Insert(PRODUCT_DB).Rows(
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

	product, err = ProductRepositoryPtr.GetProductById(uuid, exec...)
	return product, err
}

func (ProductRepositoryPtr *ProductRepository) UpdateProduct(id uuid.UUID, preUpdateProduct models.Product, exec ...utils.Executor) (models.Product, error) {

	db := utils.ResolveExecutor(ProductRepositoryPtr.Db, exec)

	_, err := db.Update(PRODUCT_DB).Set(
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

func (ProductRepositoryPtr *ProductRepository) DeleteProduct(id uuid.UUID, exec ...utils.Executor) error {
	db := utils.ResolveExecutor(ProductRepositoryPtr.Db, exec)

	_, err := db.Update(PRODUCT_DB).Set(
		goqu.Record{
			"deleted_at": time.Now(),
		},
	).Where(
		goqu.C("id").Eq(id),
		goqu.C("deleted_at").IsNull(),
	).Executor().Exec()

	return err
}
