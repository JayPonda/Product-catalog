package services

import (
	"errors"

	"github.com/lib/pq"
)

var ErrDuplicateProductName = errors.New("product with this name already exists")
var ErrDuplicateCategoryName = errors.New("category with this name already exists")
var ErrEmptyCategoryName = errors.New("category name cannot be empty")
var ErrProductNotFound = errors.New("product not found")
var ErrCategoryNotFound = errors.New("category not found")

var ErrDuplicateEmail = errors.New("email already registered")
var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrRefreshTokenNotFound = errors.New("refresh token not found or expired")

const uqProductsNameActive = "uq_products_name_active"
const uqCategoriesNameActive = "uq_categories_name_active"
const uqUsersEmailActive = "uq_users_email_active"

func IsDuplicateProductName(err error) bool {
	return isUniqueViolation(err, uqProductsNameActive)
}

func IsDuplicateCategoryName(err error) bool {
	return isUniqueViolation(err, uqCategoriesNameActive)
}

func IsDuplicateEmail(err error) bool {
	return isUniqueViolation(err, uqUsersEmailActive)
}

func isUniqueViolation(err error, indexName string) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && pqErr.Constraint == indexName
	}
	return false
}
