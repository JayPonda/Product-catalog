package services

import (
	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/google/uuid"
)

type CategoryService struct {
	Logger          *utils.StructuredLogger
	CategoryManager *repositories.CategoryRepository
}

func InitCategoryService(
	logger *utils.StructuredLogger,
	categoryManager *repositories.CategoryRepository,
) (*CategoryService, error) {
	return &CategoryService{
		Logger:          logger,
		CategoryManager: categoryManager,
	}, nil
}

func (categoryServicePtr *CategoryService) GetCategoryById(id uuid.UUID) (models.Category, error) {
	return categoryServicePtr.CategoryManager.GetCategoryById(id)
}

func (categoryServicePtr *CategoryService) GetCategoryByNames(names []string) ([]models.Category, error) {
	return categoryServicePtr.CategoryManager.GetCategoryByNames(names)
}

func (categoryServicePtr *CategoryService) ListCategories(limit int, offset int) (v1.ListCategoriesResponse, error) {
	var response v1.ListCategoriesResponse

	categories, total, err := categoryServicePtr.CategoryManager.GetCategories(limit, offset)
	if err != nil {
		categoryServicePtr.Logger.Error("CategoryService.go", "ListCategories", "failed to list categories", utils.LoggerMeta{"limit": limit, "offset": offset}, err.Error())
		return response, err
	}

	response.Categories = categories
	response.Total = total
	response.Limit = limit
	response.Offset = offset

	categoryServicePtr.Logger.Debug("CategoryService.go", "ListCategories", "categories listed", utils.LoggerMeta{"count": len(categories), "total": total})
	return response, nil
}

func (categoryServicePtr *CategoryService) MatchCategories(prefix string, limit int) ([]models.Category, error) {
	var result, err = categoryServicePtr.CategoryManager.MatchCategoriesByName(prefix, limit)

	return result, err

}

func (categoryServicePtr *CategoryService) CreateCategory(
	category models.Category,
) (models.Category, error) {

	name := utils.NormalizeName(category.Name)
	if name == "" {
		categoryServicePtr.Logger.Warn("CategoryService.go", "CreateCategory", "empty category name", utils.LoggerMeta{})
		return models.Category{}, ErrEmptyCategoryName
	}

	existing, err := categoryServicePtr.GetCategoryByNames([]string{name})
	if err != nil {
		categoryServicePtr.Logger.Error("CategoryService.go", "CreateCategory", "failed to check existing categories", utils.LoggerMeta{"name": name}, err.Error())
		return models.Category{}, err
	}
	if len(existing) > 0 {
		categoryServicePtr.Logger.Warn("CategoryService.go", "CreateCategory", "duplicate category name", utils.LoggerMeta{"name": name})
		return models.Category{}, ErrDuplicateCategoryName
	}

	created, err := categoryServicePtr.CategoryManager.CreateCategory(models.Category{Name: name})
	if err != nil {
		if IsDuplicateCategoryName(err) {
			categoryServicePtr.Logger.Warn("CategoryService.go", "CreateCategory", "duplicate category name on insert", utils.LoggerMeta{"name": name})
			return models.Category{}, ErrDuplicateCategoryName
		}
		categoryServicePtr.Logger.Error("CategoryService.go", "CreateCategory", "failed to create category", utils.LoggerMeta{"name": name}, err.Error())
		return models.Category{}, err
	}

	categoryServicePtr.Logger.Debug("CategoryService.go", "CreateCategory", "category created", utils.LoggerMeta{"id": created.ID.String(), "name": name})
	return created, nil
}

func (categoryServicePtr *CategoryService) DeleteCategory(id uuid.UUID) error {
	err := categoryServicePtr.CategoryManager.DeleteCategory(id)
	if err != nil {
		categoryServicePtr.Logger.Error("CategoryService.go", "DeleteCategory", "failed to delete category", utils.LoggerMeta{"id": id.String()}, err.Error())
		return err
	}
	categoryServicePtr.Logger.Debug("CategoryService.go", "DeleteCategory", "category deleted", utils.LoggerMeta{"id": id.String()})
	return nil
}
