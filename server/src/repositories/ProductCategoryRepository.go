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
	ctx utils.RequestContext,
	productID uuid.UUID,
	exec ...utils.Executor,
) ([]models.ProductCategory, error) {
	var productCategories []models.ProductCategory
	l := repository.Logger

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
		l.Error(ctx, "ProductCategoryRepository.go", "GetCategoriesByProduct", "failed to query product categories", utils.LoggerMeta{"product_id": productID.String()}, err.Error())
		return nil, err
	}

	return productCategories, nil
}

func (repository *ProductCategoryRepository) GetCategoriesByProductIds(
	ctx utils.RequestContext,
	productIDs []uuid.UUID,
	exec ...utils.Executor,
) ([]models.ProductCategory, error) {
	if len(productIDs) == 0 {
		return []models.ProductCategory{}, nil
	}

	var productCategories []models.ProductCategory
	l := repository.Logger

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
		l.Error(ctx, "ProductCategoryRepository.go", "GetCategoriesByProductIds", "failed to query product categories by product ids", utils.LoggerMeta{"count": len(productIDs)}, err.Error())
		return nil, err
	}

	return productCategories, nil
}

// GetProductCategory returns the link row for a product-category pair,
// regardless of whether it is currently active.
func (repository *ProductCategoryRepository) GetProductCategory(
	ctx utils.RequestContext,
	productID uuid.UUID,
	categoryID uuid.UUID,
	exec ...utils.Executor,
) (models.ProductCategory, bool, error) {
	var productCategory models.ProductCategory
	l := repository.Logger

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

	if err != nil {
		l.Error(ctx, "ProductCategoryRepository.go", "GetProductCategory", "failed to query product category", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return models.ProductCategory{}, false, err
	}

	if !found {
		return models.ProductCategory{}, false, nil
	}

	return productCategory, true, nil
}

// LinkCategory creates a product-category relationship.
// A previously soft-deleted relationship is reactivated.
func (repository *ProductCategoryRepository) LinkCategory(
	ctx utils.RequestContext,
	productID uuid.UUID,
	categoryID uuid.UUID,
	exec ...utils.Executor,
) error {
	l := repository.Logger

	db := utils.ResolveExecutor(repository.Db, exec)

	existing, found, err := repository.GetProductCategory(ctx, productID, categoryID, exec...)
	if err != nil {
		l.Error(ctx, "ProductCategoryRepository.go", "LinkCategory", "failed to check existing product category", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
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
			if err != nil {
				l.Error(ctx, "ProductCategoryRepository.go", "LinkCategory", "failed to reactivate product category link", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
				return err
			}
			l.Debug(ctx, "ProductCategoryRepository.go", "LinkCategory", "product category link reactivated", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()})
		}
		return nil
	}

	id, err := utils.GetUUID()
	if err != nil {
		l.Error(ctx, "ProductCategoryRepository.go", "LinkCategory", "failed to generate UUID", nil, err.Error())
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

	if err != nil {
		l.Error(ctx, "ProductCategoryRepository.go", "LinkCategory", "failed to insert product category link", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return err
	}

	l.Debug(ctx, "ProductCategoryRepository.go", "LinkCategory", "product category link created", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()})
	return nil
}

// UnlinkCategory soft deletes an active product-category relationship.
func (repository *ProductCategoryRepository) UnlinkCategory(
	ctx utils.RequestContext,
	productID uuid.UUID,
	categoryID uuid.UUID,
	exec ...utils.Executor,
) error {
	l := repository.Logger

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

	if err != nil {
		l.Error(ctx, "ProductCategoryRepository.go", "UnlinkCategory", "failed to unlink product category", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return err
	}

	l.Debug(ctx, "ProductCategoryRepository.go", "UnlinkCategory", "product category unlinked", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()})
	return nil
}

// DeleteProductCategories soft deletes all category relationships
// belonging to a product.
func (repository *ProductCategoryRepository) DeleteProductCategories(
	ctx utils.RequestContext,
	productID uuid.UUID,
	exec ...utils.Executor,
) error {
	l := repository.Logger

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

	if err != nil {
		l.Error(ctx, "ProductCategoryRepository.go", "DeleteProductCategories", "failed to delete product categories", utils.LoggerMeta{"product_id": productID.String()}, err.Error())
		return err
	}

	l.Debug(ctx, "ProductCategoryRepository.go", "DeleteProductCategories", "product categories deleted", utils.LoggerMeta{"product_id": productID.String()})
	return nil
}
