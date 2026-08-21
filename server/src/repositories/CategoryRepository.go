package repositories

import (
	"database/sql"
	"strings"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

type CategoryRepository struct {
	Db     *goqu.Database
	Logger *utils.StructuredLogger
}

const CATEGORY_DB = "categories"

func InitCategoryRepository(
	db *goqu.Database,
	logger *utils.StructuredLogger,
) (*CategoryRepository, error) {
	return &CategoryRepository{
		Db:     db,
		Logger: logger,
	}, nil
}

func (CategoryRepositoryPtr *CategoryRepository) GetCategoryById(id uuid.UUID, exec ...utils.Executor) (models.Category, error) {
	var category models.Category

	db := utils.ResolveExecutor(CategoryRepositoryPtr.Db, exec)

	found, err := db.
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

func (CategoryRepositoryPtr *CategoryRepository) GetCategoryByIds(ids []uuid.UUID, exec ...utils.Executor) ([]models.Category, error) {
	var category []models.Category

	db := utils.ResolveExecutor(CategoryRepositoryPtr.Db, exec)

	err := db.
		From(CATEGORY_DB).
		Where(
			goqu.C("id").In(ids),
			goqu.C("deleted_at").IsNull(),
		).
		Select(
			"id",
			"name",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStructs(&category)

	if err != nil {
		return category, err
	}

	return category, nil
}

func (CategoryRepositoryPtr *CategoryRepository) GetCategoryByNames(categories []string, exec ...utils.Executor) ([]models.Category, error) {
	var category []models.Category

	normalized := make([]string, 0, len(categories))
	for _, name := range categories {
		if name = utils.NormalizeName(name); name != "" {
			normalized = append(normalized, name)
		}
	}

	db := utils.ResolveExecutor(CategoryRepositoryPtr.Db, exec)

	err := db.
		From(CATEGORY_DB).
		Where(
			goqu.C("name").In(normalized),
			goqu.C("deleted_at").IsNull(),
		).
		Select(
			"id",
			"name",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStructs(&category)

	if err != nil {
		return nil, err
	}

	return category, nil
}

func (CategoryRepositoryPtr *CategoryRepository) GetCategories(limit int, offset int, exec ...utils.Executor) ([]models.Category, int64, error) {
	db := utils.ResolveExecutor(CategoryRepositoryPtr.Db, exec)

	var categories []models.Category

	err := db.
		From(CATEGORY_DB).
		Where(
			goqu.C("deleted_at").IsNull(),
		).
		Order(goqu.I("name").Asc()).
		Limit(uint(limit)).
		Offset(uint(offset)).
		Select(
			"id",
			"name",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStructs(&categories)

	if err != nil {
		return nil, 0, err
	}

	var total int64

	_, err = db.
		From(CATEGORY_DB).
		Where(
			goqu.C("deleted_at").IsNull(),
		).
		Select(goqu.COUNT("*")).
		ScanVal(&total)

	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (CategoryRepositoryPtr *CategoryRepository) CreateCategory(
	category models.Category,
	exec ...utils.Executor,
) (models.Category, error) {

	id, err := utils.GetUUID()
	if err != nil {
		return category, err
	}

	db := utils.ResolveExecutor(CategoryRepositoryPtr.Db, exec)

	_, err = db.
		Insert(CATEGORY_DB).
		Rows(
			goqu.Record{
				"id":   id,
				"name": utils.NormalizeName(category.Name),
			},
		).
		Executor().
		Exec()

	if err != nil {
		return category, err
	}

	return CategoryRepositoryPtr.GetCategoryById(id, exec...)
}

func (CategoryRepositoryPtr *CategoryRepository) MatchCategoriesByName(prefix string, limit int, exec ...utils.Executor) ([]models.Category, error) {
	// initialized (not nil) so zero matches marshal to [] instead of null
	categories := []models.Category{}

	escaper := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	pattern := escaper.Replace(utils.NormalizeName(prefix)) + "%"

	db := utils.ResolveExecutor(CategoryRepositoryPtr.Db, exec)

	err := db.
		From(CATEGORY_DB).
		Where(
			goqu.C("deleted_at").IsNull(),
			goqu.L("LOWER(name) LIKE ?", pattern),
		).
		Order(goqu.I("name").Asc()).
		Limit(uint(limit)).
		Select(
			"id",
			"name",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStructs(&categories)

	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (CategoryRepositoryPtr *CategoryRepository) DeleteCategory(id uuid.UUID, exec ...utils.Executor) error {
	db := utils.ResolveExecutor(CategoryRepositoryPtr.Db, exec)

	_, err := db.
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
