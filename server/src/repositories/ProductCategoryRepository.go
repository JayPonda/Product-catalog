package repositories

import (
	"time"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

type ProductCategoryRepository struct {
	Db     *goqu.Database
	Logger *utils.StructuredLogger
}

const PRODUCT_CATEGORY_DB = "product_categories"

func InitProductCategoryRepository(
	db *goqu.Database,
	logger *utils.StructuredLogger,
) (*ProductCategoryRepository, error) {
	return &ProductCategoryRepository{
		Db:     db,
		Logger: logger,
	}, nil
}

// GetCategoriesByProduct returns all active category relationships
// for a product.
func (repository *ProductCategoryRepository) GetCategoriesByProduct(
	productID uuid.UUID,
	exec ...utils.Executor,
) ([]models.ProductCategory, error) {
	var productCategories []models.ProductCategory

	db := utils.ResolveExecutor(repository.Db, exec)

	err := db.
		From(PRODUCT_CATEGORY_DB).
		Where(
			goqu.C("product_id").Eq(productID),
			goqu.C("deleted_at").IsNull(),
		).
		Select(
			"id",
			"product_id",
			"category_id",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStructs(&productCategories)

	if err != nil {
		return nil, err
	}

	return productCategories, nil
}

func (repository *ProductCategoryRepository) GetCategoriesByProductIds(
	productIDs []uuid.UUID,
	exec ...utils.Executor,
) ([]models.ProductCategory, error) {
	var productCategories []models.ProductCategory

	db := utils.ResolveExecutor(repository.Db, exec)

	err := db.
		From(PRODUCT_CATEGORY_DB).
		Where(
			goqu.C("product_id").In(productIDs),
			goqu.C("deleted_at").IsNull(),
		).
		Select(
			"id",
			"product_id",
			"category_id",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStructs(&productCategories)

	if err != nil {
		return nil, err
	}

	return productCategories, nil
}

// GetProductCategory returns the link row for a product-category pair,
// regardless of whether it is currently active.
func (repository *ProductCategoryRepository) GetProductCategory(
	productID uuid.UUID,
	categoryID uuid.UUID,
	exec ...utils.Executor,
) (models.ProductCategory, bool, error) {
	var productCategory models.ProductCategory

	db := utils.ResolveExecutor(repository.Db, exec)

	found, err := db.
		From(PRODUCT_CATEGORY_DB).
		Where(
			goqu.C("product_id").Eq(productID),
			goqu.C("category_id").Eq(categoryID),
		).
		Select(
			"id",
			"product_id",
			"category_id",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStruct(&productCategory)

	if err != nil || !found {
		return models.ProductCategory{}, false, err
	}

	return productCategory, true, nil
}

// LinkCategory creates a product-category relationship.
// A previously soft-deleted relationship is reactivated.
func (repository *ProductCategoryRepository) LinkCategory(
	productID uuid.UUID,
	categoryID uuid.UUID,
	exec ...utils.Executor,
) error {
	db := utils.ResolveExecutor(repository.Db, exec)

	existing, found, err := repository.GetProductCategory(productID, categoryID, exec...)
	if err != nil {
		return err
	}

	if found {
		if existing.DeletedAt.Valid {
			_, err = db.
				Update(PRODUCT_CATEGORY_DB).
				Set(
					goqu.Record{
						"deleted_at": nil,
						"updated_at": time.Now(),
					},
				).
				Where(goqu.C("id").Eq(existing.ID)).
				Executor().
				Exec()
		}
		return err
	}

	id, err := utils.GetUUID()
	if err != nil {
		return err
	}

	_, err = db.
		Insert(PRODUCT_CATEGORY_DB).
		Rows(
			goqu.Record{
				"id":          id,
				"product_id":  productID,
				"category_id": categoryID,
			},
		).
		Executor().
		Exec()

	return err
}

// UnlinkCategory soft deletes an active product-category relationship.
func (repository *ProductCategoryRepository) UnlinkCategory(
	productID uuid.UUID,
	categoryID uuid.UUID,
	exec ...utils.Executor,
) error {
	db := utils.ResolveExecutor(repository.Db, exec)

	_, err := db.
		Update(PRODUCT_CATEGORY_DB).
		Set(
			goqu.Record{
				"deleted_at": time.Now(),
			},
		).
		Where(
			goqu.C("product_id").Eq(productID),
			goqu.C("category_id").Eq(categoryID),
			goqu.C("deleted_at").IsNull(),
		).
		Executor().
		Exec()

	return err
}

// DeleteProductCategories soft deletes all category relationships
// belonging to a product.
func (repository *ProductCategoryRepository) DeleteProductCategories(
	productID uuid.UUID,
	exec ...utils.Executor,
) error {
	db := utils.ResolveExecutor(repository.Db, exec)

	_, err := db.
		Update(PRODUCT_CATEGORY_DB).
		Set(
			goqu.Record{
				"deleted_at": time.Now(),
			},
		).
		Where(
			goqu.C("product_id").Eq(productID),
			goqu.C("deleted_at").IsNull(),
		).
		Executor().
		Exec()

	return err
}
