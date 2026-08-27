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
		if errors.Is(err, sql.ErrNoRows) {
			productServicePtr.Logger.Warn("ProductService.go", "GetProductById", "product not found", utils.LoggerMeta{"id": id.String()})
		} else {
			productServicePtr.Logger.Error("ProductService.go", "GetProductById", "failed to get product", utils.LoggerMeta{"id": id.String()}, err.Error())
		}
		return responseProduct, err
	}

	categories, err := productServicePtr.getProductsCategory(id)
	if err != nil {
		productServicePtr.Logger.Error("ProductService.go", "GetProductById", "failed to get product categories", utils.LoggerMeta{"id": id.String()}, err.Error())
		return responseProduct, err
	}

	responseProduct.Product = products
	responseProduct.Categories = categories
	productServicePtr.Logger.Debug("ProductService.go", "GetProductById", "product retrieved", utils.LoggerMeta{"id": id.String(), "name": products.Name})
	return responseProduct, nil
}

func (productServicePtr *ProductService) GetProductByName(name string) (v1.ResponseProduct, error) {

	var responseProduct v1.ResponseProduct
	products, err := productServicePtr.ProductManager.GetProductByName(name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			productServicePtr.Logger.Warn("ProductService.go", "GetProductByName", "product not found", utils.LoggerMeta{"name": name})
		} else {
			productServicePtr.Logger.Error("ProductService.go", "GetProductByName", "failed to get product", utils.LoggerMeta{"name": name}, err.Error())
		}
		return responseProduct, err
	}

	categories, err := productServicePtr.getProductsCategory(products.ID)
	if err != nil {
		productServicePtr.Logger.Error("ProductService.go", "GetProductByName", "failed to get product categories", utils.LoggerMeta{"name": name}, err.Error())
		return responseProduct, err
	}

	responseProduct.Product = products
	responseProduct.Categories = categories
	productServicePtr.Logger.Debug("ProductService.go", "GetProductByName", "product retrieved", utils.LoggerMeta{"id": products.ID.String(), "name": name})
	return responseProduct, nil

}

func (productServicePtr *ProductService) ListProducts(limit int, offset int) (v1.ListProductsResponse, error) {
	var response v1.ListProductsResponse

	products, total, err := productServicePtr.ProductManager.GetProducts(limit, offset)
	if err != nil {
		productServicePtr.Logger.Error("ProductService.go", "ListProducts", "failed to list products", utils.LoggerMeta{"limit": limit, "offset": offset}, err.Error())
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
			productServicePtr.Logger.Error("ProductService.go", "ListProducts", "failed to get product category links", utils.LoggerMeta{"product_count": len(productIDs)}, err.Error())
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
			productServicePtr.Logger.Error("ProductService.go", "ListProducts", "failed to get categories", utils.LoggerMeta{"category_count": len(categoryIDs)}, err.Error())
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

	productServicePtr.Logger.Debug("ProductService.go", "ListProducts", "products listed", utils.LoggerMeta{"count": len(items), "total": total})
	return response, nil
}

func (productServicePtr *ProductService) CreateProduct(product v1.RequestProduct, userID uuid.UUID) (v1.ResponseProduct, error) {
	_, err := productServicePtr.ProductManager.GetProductByName(product.Name)
	if err == nil {
		productServicePtr.Logger.Warn("ProductService.go", "CreateProduct", "duplicate product name", utils.LoggerMeta{"name": product.Name})
		return v1.ResponseProduct{}, ErrDuplicateProductName
	}
	if !errors.Is(err, sql.ErrNoRows) {
		productServicePtr.Logger.Error("ProductService.go", "CreateProduct", "failed to check duplicate product name", utils.LoggerMeta{"name": product.Name}, err.Error())
		return v1.ResponseProduct{}, err
	}

	tx, err := productServicePtr.Db.Begin()
	if err != nil {
		productServicePtr.Logger.Error("ProductService.go", "CreateProduct", "failed to begin transaction", utils.LoggerMeta{"name": product.Name}, err.Error())
		return v1.ResponseProduct{}, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	dbProduct, err := productServicePtr.ProductManager.CreateProduct(models.Product{
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		StockQuantity: product.StockQuantity,
		UserID:        uuid.NullUUID{Valid: true, UUID: userID},
	}, tx)
	if err != nil {
		if IsDuplicateProductName(err) {
			productServicePtr.Logger.Warn("ProductService.go", "CreateProduct", "duplicate product name on insert", utils.LoggerMeta{"name": product.Name})
			return v1.ResponseProduct{}, ErrDuplicateProductName
		}
		productServicePtr.Logger.Error("ProductService.go", "CreateProduct", "failed to create product", utils.LoggerMeta{"name": product.Name}, err.Error())
		return v1.ResponseProduct{}, err
	}

	if err := tx.Commit(); err != nil {
		productServicePtr.Logger.Error("ProductService.go", "CreateProduct", "failed to commit transaction", utils.LoggerMeta{"name": product.Name}, err.Error())
		return v1.ResponseProduct{}, err
	}

	productServicePtr.Logger.Debug("ProductService.go", "CreateProduct", "product created", utils.LoggerMeta{"id": dbProduct.ID.String(), "name": product.Name})
	return productServicePtr.GetProductById(dbProduct.ID)
}

// ListMyProducts returns the products owned by the given user.
func (productServicePtr *ProductService) ListMyProducts(userID uuid.UUID, limit int, offset int) (v1.ListProductsResponse, error) {
	var response v1.ListProductsResponse

	products, total, err := productServicePtr.ProductManager.GetMyProducts(userID, limit, offset)
	if err != nil {
		productServicePtr.Logger.Error("ProductService.go", "ListMyProducts", "failed to list user products", utils.LoggerMeta{"user_id": userID.String(), "limit": limit, "offset": offset}, err.Error())
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
			productServicePtr.Logger.Error("ProductService.go", "ListMyProducts", "failed to get product category links", utils.LoggerMeta{"user_id": userID.String(), "product_count": len(productIDs)}, err.Error())
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
			productServicePtr.Logger.Error("ProductService.go", "ListMyProducts", "failed to get categories", utils.LoggerMeta{"user_id": userID.String(), "category_count": len(categoryIDs)}, err.Error())
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

	productServicePtr.Logger.Debug("ProductService.go", "ListMyProducts", "user products listed", utils.LoggerMeta{"user_id": userID.String(), "count": len(items), "total": total})
	return response, nil
}

func (productServicePtr *ProductService) UpdateProduct(id uuid.UUID, product v1.RequestProduct) (v1.ResponseProduct, error) {
	existing, err := productServicePtr.ProductManager.GetProductByName(product.Name)
	if err == nil && existing.ID != id {
		productServicePtr.Logger.Warn("ProductService.go", "UpdateProduct", "duplicate product name", utils.LoggerMeta{"id": id.String(), "name": product.Name})
		return v1.ResponseProduct{}, ErrDuplicateProductName
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		productServicePtr.Logger.Error("ProductService.go", "UpdateProduct", "failed to check duplicate product name", utils.LoggerMeta{"id": id.String(), "name": product.Name}, err.Error())
		return v1.ResponseProduct{}, err
	}

	tx, err := productServicePtr.Db.Begin()
	if err != nil {
		productServicePtr.Logger.Error("ProductService.go", "UpdateProduct", "failed to begin transaction", utils.LoggerMeta{"id": id.String()}, err.Error())
		return v1.ResponseProduct{}, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	_, err = productServicePtr.ProductManager.UpdateProduct(id, models.Product{
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		StockQuantity: product.StockQuantity,
	}, tx)
	if err != nil {
		if IsDuplicateProductName(err) {
			productServicePtr.Logger.Warn("ProductService.go", "UpdateProduct", "duplicate product name on update", utils.LoggerMeta{"id": id.String(), "name": product.Name})
			return v1.ResponseProduct{}, ErrDuplicateProductName
		}
		productServicePtr.Logger.Error("ProductService.go", "UpdateProduct", "failed to update product", utils.LoggerMeta{"id": id.String(), "name": product.Name}, err.Error())
		return v1.ResponseProduct{}, err
	}

	if err := tx.Commit(); err != nil {
		productServicePtr.Logger.Error("ProductService.go", "UpdateProduct", "failed to commit transaction", utils.LoggerMeta{"id": id.String()}, err.Error())
		return v1.ResponseProduct{}, err
	}

	productServicePtr.Logger.Debug("ProductService.go", "UpdateProduct", "product updated", utils.LoggerMeta{"id": id.String(), "name": product.Name})
	return productServicePtr.GetProductById(id)
}

// LinkCategory links an existing category to a product.
// Categories are created separately via the category routes.
func (productServicePtr *ProductService) LinkCategory(productID uuid.UUID, categoryID uuid.UUID) (v1.ResponseProduct, error) {
	var response v1.ResponseProduct

	if _, err := productServicePtr.ProductManager.GetProductById(productID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			productServicePtr.Logger.Warn("ProductService.go", "LinkCategory", "product not found", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()})
			return response, ErrProductNotFound
		}
		productServicePtr.Logger.Error("ProductService.go", "LinkCategory", "failed to get product", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return response, err
	}

	if _, err := productServicePtr.CategoryManager.GetCategoryById(categoryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			productServicePtr.Logger.Warn("ProductService.go", "LinkCategory", "category not found", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()})
			return response, ErrCategoryNotFound
		}
		productServicePtr.Logger.Error("ProductService.go", "LinkCategory", "failed to get category", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return response, err
	}

	tx, err := productServicePtr.Db.Begin()
	if err != nil {
		productServicePtr.Logger.Error("ProductService.go", "LinkCategory", "failed to begin transaction", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return response, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := productServicePtr.ProductCategoryManager.LinkCategory(productID, categoryID, tx); err != nil {
		productServicePtr.Logger.Error("ProductService.go", "LinkCategory", "failed to link category", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return response, err
	}

	if err := tx.Commit(); err != nil {
		productServicePtr.Logger.Error("ProductService.go", "LinkCategory", "failed to commit transaction", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return response, err
	}

	productServicePtr.Logger.Debug("ProductService.go", "LinkCategory", "category linked to product", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()})
	return productServicePtr.GetProductById(productID)
}

// UnlinkCategory removes a category link from a product.
func (productServicePtr *ProductService) UnlinkCategory(productID uuid.UUID, categoryID uuid.UUID) (v1.ResponseProduct, error) {
	var response v1.ResponseProduct

	if _, err := productServicePtr.ProductManager.GetProductById(productID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			productServicePtr.Logger.Warn("ProductService.go", "UnlinkCategory", "product not found", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()})
			return response, ErrProductNotFound
		}
		productServicePtr.Logger.Error("ProductService.go", "UnlinkCategory", "failed to get product", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return response, err
	}

	if _, err := productServicePtr.CategoryManager.GetCategoryById(categoryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			productServicePtr.Logger.Warn("ProductService.go", "UnlinkCategory", "category not found", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()})
			return response, ErrCategoryNotFound
		}
		productServicePtr.Logger.Error("ProductService.go", "UnlinkCategory", "failed to get category", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return response, err
	}

	tx, err := productServicePtr.Db.Begin()
	if err != nil {
		productServicePtr.Logger.Error("ProductService.go", "UnlinkCategory", "failed to begin transaction", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return response, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := productServicePtr.ProductCategoryManager.UnlinkCategory(productID, categoryID, tx); err != nil {
		productServicePtr.Logger.Error("ProductService.go", "UnlinkCategory", "failed to unlink category", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return response, err
	}

	if err := tx.Commit(); err != nil {
		productServicePtr.Logger.Error("ProductService.go", "UnlinkCategory", "failed to commit transaction", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()}, err.Error())
		return response, err
	}

	productServicePtr.Logger.Debug("ProductService.go", "UnlinkCategory", "category unlinked from product", utils.LoggerMeta{"product_id": productID.String(), "category_id": categoryID.String()})
	return productServicePtr.GetProductById(productID)
}

func (productServicePtr *ProductService) DeleteProduct(id uuid.UUID) error {
	tx, err := productServicePtr.Db.Begin()
	if err != nil {
		productServicePtr.Logger.Error("ProductService.go", "DeleteProduct", "failed to begin transaction", utils.LoggerMeta{"id": id.String()}, err.Error())
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	err = productServicePtr.ProductManager.DeleteProduct(id, tx)
	if err != nil {
		productServicePtr.Logger.Error("ProductService.go", "DeleteProduct", "failed to delete product", utils.LoggerMeta{"id": id.String()}, err.Error())
		return err
	}

	err = productServicePtr.ProductCategoryManager.DeleteProductCategories(id, tx)
	if err != nil {
		productServicePtr.Logger.Error("ProductService.go", "DeleteProduct", "failed to delete product categories", utils.LoggerMeta{"id": id.String()}, err.Error())
		return err
	}

	if err := tx.Commit(); err != nil {
		productServicePtr.Logger.Error("ProductService.go", "DeleteProduct", "failed to commit transaction", utils.LoggerMeta{"id": id.String()}, err.Error())
		return err
	}

	productServicePtr.Logger.Debug("ProductService.go", "DeleteProduct", "product deleted", utils.LoggerMeta{"id": id.String()})
	return nil
}
