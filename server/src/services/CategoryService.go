package services

import (
	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
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

func (categoryServicePtr *CategoryService) MatchCategories(prefix string, limit int) ([]models.Category, error) {
	return categoryServicePtr.CategoryManager.MatchCategoriesByName(prefix, limit)
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
