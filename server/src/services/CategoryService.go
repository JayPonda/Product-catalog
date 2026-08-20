package services

import (
	"database/sql"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

type CategoryService struct {
	Db     *goqu.Database
	Logger *utils.StructuredLogger
}

const CATEGORY_DB = "categories"

func InitCategoryService(
	db *goqu.Database,
	logger *utils.StructuredLogger,
) (*CategoryService, error) {
	return &CategoryService{
		Db:     db,
		Logger: logger,
	}, nil
}

func (categoryServicePtr *CategoryService) GetCategoryById(id uuid.UUID) (models.Category, error) {
	var category models.Category

	found, err := categoryServicePtr.Db.
		From(CATEGORY_DB).
		Where(
			goqu.C("id").Eq(id),
			goqu.C("deleted_at").IsNull(),
		).
		Select(
			"id",
			"name",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStruct(&category)

	if err != nil {
		return category, err
	}

	if !found {
		return category, sql.ErrNoRows
	}

	return category, nil
}

func (categoryServicePtr *CategoryService) GetCategoryByNames(categories []string) ([]models.Category, error) {
	var category []models.Category

	found, err := categoryServicePtr.Db.
		From(CATEGORY_DB).
		Where(
			goqu.C("name").In(categories),
			goqu.C("deleted_at").IsNull(),
		).
		Select(
			"id",
			"name",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStruct(&category)

	if err != nil {
		return category, err
	}

	if !found {
		return category, sql.ErrNoRows
	}

	return category, nil
}

func (categoryServicePtr *CategoryService) CreateCategory(
	category models.Category,
) (models.Category, error) {

	id, err := utils.GetUUID()
	if err != nil {
		return category, err
	}

	_, err = categoryServicePtr.Db.
		Insert(CATEGORY_DB).
		Rows(
			goqu.Record{
				"id":   id,
				"name": category.Name,
			},
		).
		Executor().
		Exec()

	if err != nil {
		return category, err
	}

	return categoryServicePtr.GetCategoryById(id)
}

func (categoryServicePtr *CategoryService) UpdateCategory(
	id uuid.UUID,
	preUpdateCategory models.Category,
) (models.Category, error) {

	_, err := categoryServicePtr.Db.
		Update(CATEGORY_DB).
		Set(
			goqu.Record{
				"name":       preUpdateCategory.Name,
				"updated_at": time.Now(),
			},
		).
		Where(
			goqu.C("id").Eq(id),
			goqu.C("deleted_at").IsNull(),
		).
		Executor().
		Exec()

	if err != nil {
		return preUpdateCategory, err
	}

	return categoryServicePtr.GetCategoryById(id)
}

func (categoryServicePtr *CategoryService) DeleteCategory(id uuid.UUID) error {
	_, err := categoryServicePtr.Db.
		Update(CATEGORY_DB).
		Set(
			goqu.Record{
				"deleted_at": time.Now(),
			},
		).
		Where(
			goqu.C("id").Eq(id),
			goqu.C("deleted_at").IsNull(),
		).
		Executor().
		Exec()

	return err
}
