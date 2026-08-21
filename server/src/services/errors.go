package services

import (
	"errors"

	"github.com/lib/pq"
)

var ErrDuplicateProductName = errors.New("product with this name already exists")
var ErrDuplicateCategoryName = errors.New("category with this name already exists")
var ErrEmptyCategoryName = errors.New("category name cannot be empty")
var ErrCategoryNotModifiable = errors.New("categories can't be modified. Please create a new category instead.")

const uqProductsNameActive = "uq_products_name_active"
const uqCategoriesNameActive = "uq_categories_name_active"

func IsDuplicateProductName(err error) bool {
	return isUniqueViolation(err, uqProductsNameActive)
}

func IsDuplicateCategoryName(err error) bool {
	return isUniqueViolation(err, uqCategoriesNameActive)
}

func isUniqueViolation(err error, indexName string) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && pqErr.Constraint == indexName
	}
	return false
}
