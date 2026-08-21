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
		return response, err
	}

	response.Categories = categories
	response.Total = total
	response.Limit = limit
	response.Offset = offset

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
		return models.Category{}, ErrEmptyCategoryName
	}

	existing, err := categoryServicePtr.GetCategoryByNames([]string{name})
	if err != nil {
		return models.Category{}, err
	}
	if len(existing) > 0 {
		return models.Category{}, ErrDuplicateCategoryName
	}

	created, err := categoryServicePtr.CategoryManager.CreateCategory(models.Category{Name: name})
	if err != nil {
		if IsDuplicateCategoryName(err) {
			return models.Category{}, ErrDuplicateCategoryName
		}
		return models.Category{}, err
	}

	return created, nil
}

func (categoryServicePtr *CategoryService) DeleteCategory(id uuid.UUID) error {
	return categoryServicePtr.CategoryManager.DeleteCategory(id)
}
