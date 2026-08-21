package controllersv1

import (
	"errors"

	"github.com/JayPonda/Product-catalog/server/src/services"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/gofiber/fiber/v3"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ProductController struct {
	Service   *services.ProductService
	Validator *validator.Validate
}

func NewProductController(service *services.ProductService) *ProductController {
	return &ProductController{
		Service:   service,
		Validator: validator.New(),
	}
}

// ListProducts godoc
// @Summary      List products
// @Description  List products with pagination, newest first
// @Tags         products
// @Produce      json
// @Param        limit   query  int  false  "Page size (20|50|100)"  default(20)
// @Param        offset  query  int  false  "Offset"             default(0)
// @Success      200     {object}  v1.ListProductsResponse
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /products [get]
func (pc *ProductController) ListProducts(ctx fiber.Ctx) error {
	var query v1.ListProductsQuery

	if err := ctx.Bind().Query(&query); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
	}

	if err := pc.Validator.Struct(query); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	if query.Limit == 0 {
		query.Limit = 20
	}

	response, err := pc.Service.ListProducts(query.Limit, query.Offset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(response)
}

// CreateProduct godoc
// @Summary      Create a product
// @Description  Create a new product with categories
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        product  body      v1.RequestProduct  true  "Product payload"
// @Success      201      {object}  v1.ResponseProduct
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /products [post]
func (pc *ProductController) CreateProduct(ctx fiber.Ctx) error {
	var req v1.RequestProduct

	if err := ctx.Bind().JSON(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if err := pc.Validator.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	product, err := pc.Service.CreateProduct(req)
	if err != nil {
		if errors.Is(err, services.ErrCategoryNotModifiable) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if errors.Is(err, services.ErrDuplicateProductName) || errors.Is(err, services.ErrDuplicateCategoryName) {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(product)
}

// GetProductById godoc
// @Summary      Get a product by ID
// @Description  Get a product by its UUID
// @Tags         products
// @Produce      json
// @Param        id   path      string  true  "Product ID"
// @Success      200  {object}  v1.ResponseProduct
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /products/{id} [get]
func (pc *ProductController) GetProductById(ctx fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	product, err := pc.Service.GetProductById(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "product not found",
		})
	}

	return ctx.JSON(product)
}

// GetProductByName godoc
// @Summary      Get a product by name
// @Description  Get a product by its name
// @Tags         products
// @Produce      json
// @Param        name   path      string  true  "Product name"
// @Success      200    {object}  v1.ResponseProduct
// @Failure      404    {object}  map[string]string
// @Router       /products/name/{name} [get]
func (pc *ProductController) GetProductByName(ctx fiber.Ctx) error {
	name := ctx.Params("name")

	product, err := pc.Service.GetProductByName(name)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "product not found",
		})
	}

	return ctx.JSON(product)
}

// UpdateProduct godoc
// @Summary      Update a product
// @Description  Update a product and its categories
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      string              true  "Product ID"
// @Param        product  body      v1.RequestProduct   true  "Product payload"
// @Success      200      {object}  v1.ResponseProduct
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /products/{id} [put]
func (pc *ProductController) UpdateProduct(ctx fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	var req v1.RequestProduct

	if err := ctx.Bind().JSON(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if err := pc.Validator.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	product, err := pc.Service.UpdateProduct(id, req)
	if err != nil {
		if errors.Is(err, services.ErrCategoryNotModifiable) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if errors.Is(err, services.ErrDuplicateProductName) || errors.Is(err, services.ErrDuplicateCategoryName) {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(product)
}

// DeleteProduct godoc
// @Summary      Delete a product
// @Description  Soft-delete a product and its category relationships
// @Tags         products
// @Param        id   path      string  true  "Product ID"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /products/{id} [delete]
func (pc *ProductController) DeleteProduct(ctx fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	err = pc.Service.DeleteProduct(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusNoContent).Send(nil)
}
