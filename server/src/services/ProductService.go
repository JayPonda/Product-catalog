package services

import (
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

func (productServicePtr *ProductService) GetProductById(id uuid.UUID) (v1.BareProduct, error) {
	var bareProduct v1.BareProduct
	products, err := productServicePtr.ProductManager.GetProductById(id)

	if err != nil {
		return bareProduct, err
	}

	categories, err := productServicePtr.getProductsCategory(id)

	bareProduct.Product = products
	bareProduct.Categories = categories
	return bareProduct, nil
}

func (productServicePtr *ProductService) GetProductByName(name string) (v1.BareProduct, error) {

	var bareProduct v1.BareProduct
	products, err := productServicePtr.ProductManager.GetProductByName(name)

	if err != nil {
		return bareProduct, err
	}

	categories, err := productServicePtr.getProductsCategory(products.ID)

	bareProduct.Product = products
	bareProduct.Categories = categories
	return bareProduct, nil

}

func (productServicePtr *ProductService) CreateProduct(product v1.RequestProduct) (v1.ResponseProduct, error) {
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
	})
	if err != nil {
		return v1.ResponseProduct{}, err
	}

	var categoryIDs []uuid.UUID

	if len(product.Category.Old) > 0 {
		var names []string
		for _, cat := range product.Category.Old {
			names = append(names, cat.Name)
		}
		existingCats, err := productServicePtr.CategoryManager.GetCategoryByNames(names)
		if err != nil {
			return v1.ResponseProduct{}, err
		}
		for _, cat := range existingCats {
			categoryIDs = append(categoryIDs, cat.ID)
		}
	}

	for _, name := range product.Category.New {
		newCat, err := productServicePtr.CategoryManager.CreateCategory(
			models.Category{Name: name},
		)
		if err != nil {
			return v1.ResponseProduct{}, err
		}
		categoryIDs = append(categoryIDs, newCat.ID)
	}

	err = productServicePtr.ProductCategoryManager.SetProductCategories(
		dbProduct.ID, categoryIDs,
	)
	if err != nil {
		return v1.ResponseProduct{}, err
	}

	if err := tx.Commit(); err != nil {
		return v1.ResponseProduct{}, err
	}

	bareProduct, err := productServicePtr.GetProductById(dbProduct.ID)
	if err != nil {
		return v1.ResponseProduct{}, err
	}

	return v1.ResponseProduct{
		ID:            bareProduct.ID,
		Name:          bareProduct.Name,
		Description:   bareProduct.Description,
		Price:         bareProduct.Price,
		StockQuantity: bareProduct.StockQuantity,
		CreatedAt:     bareProduct.CreatedAt,
		UpdatedAt:     bareProduct.UpdatedAt,
		DeletedAt:     bareProduct.DeletedAt,
	}, nil
}

func (productServicePtr *ProductService) UpdateProduct(id uuid.UUID, product v1.RequestProduct) (v1.BareProduct, error) {
	tx, err := productServicePtr.Db.Begin()
	if err != nil {
		return v1.BareProduct{}, err
	}

	defer tx.Rollback()

	_, err = productServicePtr.ProductManager.UpdateProduct(id, models.Product{
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		StockQuantity: product.StockQuantity,
	})
	if err != nil {
		return v1.BareProduct{}, err
	}

	var categoryIDs []uuid.UUID

	if len(product.Category.Old) > 0 {
		var names []string
		for _, cat := range product.Category.Old {
			names = append(names, cat.Name)
		}
		existingCats, err := productServicePtr.CategoryManager.GetCategoryByNames(names)
		if err != nil {
			return v1.BareProduct{}, err
		}
		for _, cat := range existingCats {
			categoryIDs = append(categoryIDs, cat.ID)
		}
	}

	for _, name := range product.Category.New {
		newCat, err := productServicePtr.CategoryManager.CreateCategory(
			models.Category{Name: name},
		)
		if err != nil {
			return v1.BareProduct{}, err
		}
		categoryIDs = append(categoryIDs, newCat.ID)
	}

	err = productServicePtr.ProductCategoryManager.SetProductCategories(id, categoryIDs)
	if err != nil {
		return v1.BareProduct{}, err
	}

	if err := tx.Commit(); err != nil {
		return v1.BareProduct{}, err
	}

	return productServicePtr.GetProductById(id)
}

func (productServicePtr *ProductService) DeleteProduct(id uuid.UUID) error {
	tx, err := productServicePtr.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = productServicePtr.ProductManager.DeleteProduct(id)
	if err != nil {
		return err
	}

	err = productServicePtr.ProductCategoryManager.DeleteProductCategories(id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
