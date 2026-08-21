package services

import (
	"database/sql"
	"errors"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

type ProductService struct {
	Db                     *goqu.Database
	Logger                 *utils.StructuredLogger
	ProductManager         *repositories.ProductRepository
	CategoryManager        *repositories.CategoryRepository
	ProductCategoryManager *repositories.ProductCategoryRepository
}

const PRODUCT_DB = "products"

func InitProductService(db *goqu.Database, logger *utils.StructuredLogger, productManager *repositories.ProductRepository, categoryManage *repositories.CategoryRepository, productcategoryManager *repositories.ProductCategoryRepository) (*ProductService, error) {

	return &ProductService{
		Db:                     db,
		Logger:                 logger,
		ProductManager:         productManager,
		CategoryManager:        categoryManage,
		ProductCategoryManager: productcategoryManager,
	}, nil
}

func (productServicePtr *ProductService) getProductsCategory(id uuid.UUID) ([]models.Category, error) {
	productCategories, err := productServicePtr.ProductCategoryManager.GetCategoriesByProduct(id)

	if err != nil {
		return []models.Category{}, err
	}

	var categoryIds []uuid.UUID
	for _, productCategory := range productCategories {
		categoryIds = append(categoryIds, productCategory.CategoryID)
	}

	categories, err := productServicePtr.CategoryManager.GetCategoryByIds(categoryIds)
	if err != nil {
		return []models.Category{}, err
	}

	return categories, nil
}

func (productServicePtr *ProductService) GetProductById(id uuid.UUID) (v1.ResponseProduct, error) {
	var responseProduct v1.ResponseProduct
	products, err := productServicePtr.ProductManager.GetProductById(id)

	if err != nil {
		return responseProduct, err
	}

	categories, err := productServicePtr.getProductsCategory(id)

	responseProduct.Product = products
	responseProduct.Categories = categories
	return responseProduct, nil
}

func (productServicePtr *ProductService) GetProductByName(name string) (v1.ResponseProduct, error) {

	var responseProduct v1.ResponseProduct
	products, err := productServicePtr.ProductManager.GetProductByName(name)

	if err != nil {
		return responseProduct, err
	}

	categories, err := productServicePtr.getProductsCategory(products.ID)

	responseProduct.Product = products
	responseProduct.Categories = categories
	return responseProduct, nil

}

func (productServicePtr *ProductService) ListProducts(limit int, offset int) (v1.ListProductsResponse, error) {
	var response v1.ListProductsResponse

	products, total, err := productServicePtr.ProductManager.GetProducts(limit, offset)
	if err != nil {
		return response, err
	}

	productIDs := make([]uuid.UUID, 0, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}

	categoriesByProduct := make(map[uuid.UUID][]models.Category, len(products))

	if len(productIDs) > 0 {
		links, err := productServicePtr.ProductCategoryManager.GetCategoriesByProductIds(productIDs)
		if err != nil {
			return response, err
		}

		categoryIDSet := make(map[uuid.UUID]struct{}, len(links))
		for _, link := range links {
			categoryIDSet[link.CategoryID] = struct{}{}
		}

		categoryIDs := make([]uuid.UUID, 0, len(categoryIDSet))
		for categoryID := range categoryIDSet {
			categoryIDs = append(categoryIDs, categoryID)
		}

		categories, err := productServicePtr.CategoryManager.GetCategoryByIds(categoryIDs)
		if err != nil {
			return response, err
		}

		categoriesByID := make(map[uuid.UUID]models.Category, len(categories))
		for _, category := range categories {
			categoriesByID[category.ID] = category
		}

		for _, link := range links {
			if category, ok := categoriesByID[link.CategoryID]; ok {
				categoriesByProduct[link.ProductID] = append(categoriesByProduct[link.ProductID], category)
			}
		}
	}

	items := make([]v1.ResponseProduct, 0, len(products))
	for _, product := range products {
		items = append(items, v1.ResponseProduct{
			Product:    product,
			Categories: categoriesByProduct[product.ID],
		})
	}

	response.Products = items
	response.Total = total
	response.Limit = limit
	response.Offset = offset

	return response, nil
}

func (productServicePtr *ProductService) resolveCategoryIDs(tx *goqu.TxDatabase, category *v1.RequestCategories) ([]uuid.UUID, bool, error) {
	if category == nil {
		return nil, false, nil
	}
	var categoryIDs []uuid.UUID

	if len(category.Old) > 0 {
		var names []string
		for _, cat := range category.Old {
			names = append(names, cat.Name)
		}

		existingCats, err := productServicePtr.CategoryManager.GetCategoryByNames(names, tx)
		if err != nil {
			return nil, false, err
		}

		found := make(map[string]struct{}, len(existingCats))
		for _, cat := range existingCats {
			found[cat.Name] = struct{}{}
		}

		for _, name := range names {
			if _, ok := found[utils.NormalizeName(name)]; !ok {
				return nil, false, ErrCategoryNotModifiable
			}
		}

		for _, cat := range existingCats {
			categoryIDs = append(categoryIDs, cat.ID)
		}
	}

	for _, name := range category.New {
		newCat, err := productServicePtr.CategoryManager.CreateCategory(
			models.Category{Name: name},
			tx,
		)
		if err != nil {
			if IsDuplicateCategoryName(err) {
				return nil, false, ErrDuplicateCategoryName
			}
			return nil, false, err
		}
		categoryIDs = append(categoryIDs, newCat.ID)
	}

	return categoryIDs, true, nil
}

func (productServicePtr *ProductService) CreateProduct(product v1.RequestProduct) (v1.ResponseProduct, error) {
	_, err := productServicePtr.ProductManager.GetProductByName(product.Name)
	if err == nil {
		return v1.ResponseProduct{}, ErrDuplicateProductName
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return v1.ResponseProduct{}, err
	}

	tx, err := productServicePtr.Db.Begin()
	if err != nil {
		return v1.ResponseProduct{}, err
	}

	defer tx.Rollback()

	dbProduct, err := productServicePtr.ProductManager.CreateProduct(models.Product{
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		StockQuantity: product.StockQuantity,
	}, tx)
	if err != nil {
		if IsDuplicateProductName(err) {
			return v1.ResponseProduct{}, ErrDuplicateProductName
		}
		return v1.ResponseProduct{}, err
	}

	var categoryIDs []uuid.UUID

	categoryIDs, hasCategories, err := productServicePtr.resolveCategoryIDs(tx, product.Category)
	if err != nil {
		return v1.ResponseProduct{}, err
	}

	if hasCategories {
		err = productServicePtr.ProductCategoryManager.SetProductCategories(
			dbProduct.ID, categoryIDs, tx,
		)
		if err != nil {
			return v1.ResponseProduct{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return v1.ResponseProduct{}, err
	}

	return productServicePtr.GetProductById(dbProduct.ID)
}

func (productServicePtr *ProductService) UpdateProduct(id uuid.UUID, product v1.RequestProduct) (v1.ResponseProduct, error) {
	existing, err := productServicePtr.ProductManager.GetProductByName(product.Name)
	if err == nil && existing.ID != id {
		return v1.ResponseProduct{}, ErrDuplicateProductName
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return v1.ResponseProduct{}, err
	}

	tx, err := productServicePtr.Db.Begin()
	if err != nil {
		return v1.ResponseProduct{}, err
	}

	defer tx.Rollback()

	_, err = productServicePtr.ProductManager.UpdateProduct(id, models.Product{
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		StockQuantity: product.StockQuantity,
	}, tx)
	if err != nil {
		if IsDuplicateProductName(err) {
			return v1.ResponseProduct{}, ErrDuplicateProductName
		}
		return v1.ResponseProduct{}, err
	}

	categoryIDs, hasCategories, err := productServicePtr.resolveCategoryIDs(tx, product.Category)
	if err != nil {
		return v1.ResponseProduct{}, err
	}

	if hasCategories {
		err = productServicePtr.ProductCategoryManager.SetProductCategories(id, categoryIDs, tx)
		if err != nil {
			return v1.ResponseProduct{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return v1.ResponseProduct{}, err
	}

	return productServicePtr.GetProductById(id)
}

func (productServicePtr *ProductService) DeleteProduct(id uuid.UUID) error {
	tx, err := productServicePtr.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = productServicePtr.ProductManager.DeleteProduct(id, tx)
	if err != nil {
		return err
	}

	err = productServicePtr.ProductCategoryManager.DeleteProductCategories(id, tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}
