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
	l := ProductRepositoryPtr.Logger

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
			"user_id",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStruct(&product)

	if err != nil {
		l.Error("ProductRepository.go", "GetProductById", "failed to query product", utils.LoggerMeta{"id": id.String()}, err.Error())
		return product, err
	}

	if !found {
		l.Warn("ProductRepository.go", "GetProductById", "product not found", utils.LoggerMeta{"id": id.String()})
		return product, sql.ErrNoRows
	}

	return product, nil
}

func (ProductRepositoryPtr *ProductRepository) GetProductByName(name string, exec ...utils.Executor) (models.Product, error) {
	var product models.Product
	l := ProductRepositoryPtr.Logger

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
			"user_id",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStruct(&product)

	if err != nil {
		l.Error("ProductRepository.go", "GetProductByName", "failed to query product by name", utils.LoggerMeta{"name": name}, err.Error())
		return product, err
	}

	if !found {
		l.Warn("ProductRepository.go", "GetProductByName", "product not found by name", utils.LoggerMeta{"name": name})
		return product, sql.ErrNoRows
	}

	return product, nil
}

func (ProductRepositoryPtr *ProductRepository) GetProducts(limit int, offset int, exec ...utils.Executor) ([]models.Product, int64, error) {
	db := utils.ResolveExecutor(ProductRepositoryPtr.Db, exec)
	l := ProductRepositoryPtr.Logger

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
			"user_id",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStructs(&products)

	if err != nil {
		l.Error("ProductRepository.go", "GetProducts", "failed to query products", utils.LoggerMeta{"limit": limit, "offset": offset}, err.Error())
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
		l.Error("ProductRepository.go", "GetProducts", "failed to count products", utils.LoggerMeta{"limit": limit, "offset": offset}, err.Error())
		return nil, 0, err
	}

	l.Debug("ProductRepository.go", "GetProducts", "products retrieved", utils.LoggerMeta{"limit": limit, "offset": offset, "count": len(products), "total": total})
	return products, total, nil
}

func (ProductRepositoryPtr *ProductRepository) CreateProduct(product models.Product, exec ...utils.Executor) (models.Product, error) {
	l := ProductRepositoryPtr.Logger

	uuid, err := utils.GetUUID()
	if err != nil {
		l.Error("ProductRepository.go", "CreateProduct", "failed to generate UUID", nil, err.Error())
		return product, err
	}

	db := utils.ResolveExecutor(ProductRepositoryPtr.Db, exec)

	userID := interface{}(nil)
	if product.UserID.Valid {
		userID = product.UserID.UUID
	}

	_, err = db.Insert(PRODUCT_DB).Rows(
		goqu.Record{
			"id":             uuid,
			"name":           product.Name,
			"description":    product.Description,
			"price":          product.Price,
			"stock_quantity": product.StockQuantity,
			"user_id":        userID,
		},
	).Executor().Exec()

	if err != nil {
		l.Error("ProductRepository.go", "CreateProduct", "failed to insert product", utils.LoggerMeta{"name": product.Name}, err.Error())
		return product, err
	}

	product, err = ProductRepositoryPtr.GetProductById(uuid, exec...)
	if err != nil {
		l.Error("ProductRepository.go", "CreateProduct", "failed to retrieve created product", utils.LoggerMeta{"id": uuid.String()}, err.Error())
		return product, err
	}

	l.Debug("ProductRepository.go", "CreateProduct", "product created", utils.LoggerMeta{"id": uuid.String(), "name": product.Name})
	return product, nil
}

func (ProductRepositoryPtr *ProductRepository) UpdateProduct(id uuid.UUID, preUpdateProduct models.Product, exec ...utils.Executor) (models.Product, error) {
	l := ProductRepositoryPtr.Logger

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
		l.Error("ProductRepository.go", "UpdateProduct", "failed to update product", utils.LoggerMeta{"id": id.String()}, err.Error())
		return preUpdateProduct, err
	}

	product, err := ProductRepositoryPtr.GetProductById(id, exec...)
	if err != nil {
		l.Error("ProductRepository.go", "UpdateProduct", "failed to retrieve updated product", utils.LoggerMeta{"id": id.String()}, err.Error())
		return product, err
	}

	l.Debug("ProductRepository.go", "UpdateProduct", "product updated", utils.LoggerMeta{"id": id.String()})
	return product, nil
}

func (ProductRepositoryPtr *ProductRepository) DeleteProduct(id uuid.UUID, exec ...utils.Executor) error {
	l := ProductRepositoryPtr.Logger

	db := utils.ResolveExecutor(ProductRepositoryPtr.Db, exec)

	_, err := db.Update(PRODUCT_DB).Set(
		goqu.Record{
			"deleted_at": time.Now(),
		},
	).Where(
		goqu.C("id").Eq(id),
		goqu.C("deleted_at").IsNull(),
	).Executor().Exec()

	if err != nil {
		l.Error("ProductRepository.go", "DeleteProduct", "failed to delete product", utils.LoggerMeta{"id": id.String()}, err.Error())
		return err
	}

	l.Debug("ProductRepository.go", "DeleteProduct", "product deleted", utils.LoggerMeta{"id": id.String()})
	return nil
}

// GetMyProducts returns products owned by the given user with pagination.
func (ProductRepositoryPtr *ProductRepository) GetMyProducts(userID uuid.UUID, limit int, offset int, exec ...utils.Executor) ([]models.Product, int64, error) {
	db := utils.ResolveExecutor(ProductRepositoryPtr.Db, exec)
	l := ProductRepositoryPtr.Logger

	var products []models.Product

	err := db.
		From(PRODUCT_DB).
		Where(
			goqu.C("deleted_at").IsNull(),
			goqu.C("user_id").Eq(userID),
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
			"user_id",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStructs(&products)

	if err != nil {
		l.Error("ProductRepository.go", "GetMyProducts", "failed to query user products", utils.LoggerMeta{"user_id": userID.String(), "limit": limit, "offset": offset}, err.Error())
		return nil, 0, err
	}

	var total int64

	_, err = db.
		From(PRODUCT_DB).
		Where(
			goqu.C("deleted_at").IsNull(),
			goqu.C("user_id").Eq(userID),
		).
		Select(goqu.COUNT("*")).
		ScanVal(&total)

	if err != nil {
		l.Error("ProductRepository.go", "GetMyProducts", "failed to count user products", utils.LoggerMeta{"user_id": userID.String(), "limit": limit, "offset": offset}, err.Error())
		return nil, 0, err
	}

	l.Debug("ProductRepository.go", "GetMyProducts", "user products retrieved", utils.LoggerMeta{"user_id": userID.String(), "limit": limit, "offset": offset, "count": len(products), "total": total})
	return products, total, nil
}
