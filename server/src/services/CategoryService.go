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

func (categoryServicePtr *CategoryService) GetCategoryById(ctx utils.RequestContext, id uuid.UUID) (models.Category, error) {
	return categoryServicePtr.CategoryManager.GetCategoryById(ctx, id)
}

func (categoryServicePtr *CategoryService) GetCategoryByNames(ctx utils.RequestContext, names []string) ([]models.Category, error) {
	return categoryServicePtr.CategoryManager.GetCategoryByNames(ctx, names)
}

func (categoryServicePtr *CategoryService) ListCategories(ctx utils.RequestContext, limit int, offset int, filter models.CategoryFilter) (v1.ListCategoriesResponse, error) {
	var response v1.ListCategoriesResponse

	categories, total, err := categoryServicePtr.CategoryManager.GetCategories(ctx, limit, offset, filter)
	if err != nil {
		categoryServicePtr.Logger.Error(ctx, "CategoryService.go", "ListCategories", "failed to list categories", utils.LoggerMeta{"limit": limit, "offset": offset, "filter": filter.Name}, err.Error())
		return response, err
	}

	response.Categories = categories
	response.Total = total
	response.Limit = limit
	response.Offset = offset

	categoryServicePtr.Logger.Debug(ctx, "CategoryService.go", "ListCategories", "categories listed", utils.LoggerMeta{"count": len(categories), "total": total, "filter": filter.Name})
	return response, nil
}

func (categoryServicePtr *CategoryService) MatchCategories(ctx utils.RequestContext, prefix string, limit int) ([]models.Category, error) {
	var result, err = categoryServicePtr.CategoryManager.MatchCategoriesByName(ctx, prefix, limit)

	return result, err

}

func (categoryServicePtr *CategoryService) CreateCategory(
	ctx utils.RequestContext,
	category models.Category,
) (models.Category, error) {

	name := utils.NormalizeName(category.Name)
	if name == "" {
		categoryServicePtr.Logger.Warn(ctx, "CategoryService.go", "CreateCategory", "empty category name", utils.LoggerMeta{})
		return models.Category{}, ErrEmptyCategoryName
	}

	existing, err := categoryServicePtr.GetCategoryByNames(ctx, []string{name})
	if err != nil {
		categoryServicePtr.Logger.Error(ctx, "CategoryService.go", "CreateCategory", "failed to check existing categories", utils.LoggerMeta{"name": name}, err.Error())
		return models.Category{}, err
	}
	if len(existing) > 0 {
		categoryServicePtr.Logger.Warn(ctx, "CategoryService.go", "CreateCategory", "duplicate category name", utils.LoggerMeta{"name": name})
		return models.Category{}, ErrDuplicateCategoryName
	}

	created, err := categoryServicePtr.CategoryManager.CreateCategory(ctx, models.Category{Name: name})
	if err != nil {
		if IsDuplicateCategoryName(err) {
			categoryServicePtr.Logger.Warn(ctx, "CategoryService.go", "CreateCategory", "duplicate category name on insert", utils.LoggerMeta{"name": name})
			return models.Category{}, ErrDuplicateCategoryName
		}
		categoryServicePtr.Logger.Error(ctx, "CategoryService.go", "CreateCategory", "failed to create category", utils.LoggerMeta{"name": name}, err.Error())
		return models.Category{}, err
	}

	categoryServicePtr.Logger.Debug(ctx, "CategoryService.go", "CreateCategory", "category created", utils.LoggerMeta{"id": created.ID.String(), "name": name})
	return created, nil
}

func (categoryServicePtr *CategoryService) DeleteCategory(ctx utils.RequestContext, id uuid.UUID) error {
	err := categoryServicePtr.CategoryManager.DeleteCategory(ctx, id)
	if err != nil {
		categoryServicePtr.Logger.Error(ctx, "CategoryService.go", "DeleteCategory", "failed to delete category", utils.LoggerMeta{"id": id.String()}, err.Error())
		return err
	}
	categoryServicePtr.Logger.Debug(ctx, "CategoryService.go", "DeleteCategory", "category deleted", utils.LoggerMeta{"id": id.String()})
	return nil
}
