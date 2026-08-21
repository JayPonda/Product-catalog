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

// SetProductCategories makes the product-category relationships match
// the supplied category IDs.
//
// Existing relationships are kept.
// New relationships are created.
// Relationships that are no longer present are soft deleted.
func (repository *ProductCategoryRepository) SetProductCategories(
	productID uuid.UUID,
	categoryIDs []uuid.UUID,
	exec ...utils.Executor,
) error {

	db := utils.ResolveExecutor(repository.Db, exec)

	// Get current relationships.
	currentRelations, err := repository.GetCategoriesByProduct(productID, exec...)
	if err != nil {
		return err
	}

	// Build a lookup of the incoming category IDs.
	desiredCategories := make(map[uuid.UUID]struct{}, len(categoryIDs))

	for _, categoryID := range categoryIDs {
		desiredCategories[categoryID] = struct{}{}
	}

	// Build a lookup of the currently active relationships.
	currentCategories := make(map[uuid.UUID]models.ProductCategory, len(currentRelations))

	for _, relation := range currentRelations {
		currentCategories[relation.CategoryID] = relation
	}

	// Create relationships that don't currently exist.
	for categoryID := range desiredCategories {
		if _, exists := currentCategories[categoryID]; exists {
			continue
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

		if err != nil {
			return err
		}
	}

	// Soft delete relationships that are no longer desired.
	for categoryID, relation := range currentCategories {
		if _, exists := desiredCategories[categoryID]; exists {
			continue
		}

		_, err := db.
			Update(PRODUCT_CATEGORY_DB).
			Set(
				goqu.Record{
					"deleted_at": time.Now(),
				},
			).
			Where(
				goqu.C("id").Eq(relation.ID),
				goqu.C("deleted_at").IsNull(),
			).
			Executor().
			Exec()

		if err != nil {
			return err
		}
	}

	return nil
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
