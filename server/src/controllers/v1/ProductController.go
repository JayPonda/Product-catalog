package controllersv1

import (
	"errors"

	"github.com/JayPonda/Product-catalog/server/src/services"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ProductController struct {
	Service   *services.ProductService
	Validator *validator.Validate
	Logger    *utils.StructuredLogger
}

func NewProductController(service *services.ProductService, logger *utils.StructuredLogger) *ProductController {
	return &ProductController{
		Service:   service,
		Validator: utils.NewValidator(),
		Logger:    logger,
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
	pc.Logger.Debug(ctx, "ProductController.go", "ListProducts", "request received", nil)

	var query v1.ListProductsQuery

	if err := ctx.Bind().Query(&query); err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "ListProducts", "invalid query parameters", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
	}

	if err := pc.Validator.Struct(query); err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "ListProducts", "validation failed", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	if query.Limit == 0 {
		query.Limit = 20
	}

	response, err := pc.Service.ListProducts(ctx, query.Limit, query.Offset)
	if err != nil {
		pc.Logger.Error(ctx, "ProductController.go", "ListProducts", "service error", utils.LoggerMeta{"error": err.Error()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	pc.Logger.Debug(ctx, "ProductController.go", "ListProducts", "success", utils.LoggerMeta{"limit": query.Limit, "offset": query.Offset})
	return ctx.JSON(response)
}

// CreateProduct godoc
// @Summary      Create a product
// @Description  Create a new product. Categories are linked separately via the link route.
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        product  body      v1.RequestProduct  true  "Product payload"
// @Success      201      {object}  v1.ResponseProduct
// @Failure      400      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /products [post]
func (pc *ProductController) CreateProduct(ctx fiber.Ctx) error {
	pc.Logger.Debug(ctx, "ProductController.go", "CreateProduct", "request received", nil)

	var req v1.RequestProduct

	if err := ctx.Bind().JSON(&req); err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "CreateProduct", "invalid request body", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if err := pc.Validator.Struct(req); err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "CreateProduct", "validation failed", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	product, err := pc.Service.CreateProduct(ctx, req, ctx.Locals(utils.UserContextKey).(uuid.UUID))
	if err != nil {
		if errors.Is(err, services.ErrDuplicateProductName) {
			pc.Logger.Warn(ctx, "ProductController.go", "CreateProduct", "duplicate product name", utils.LoggerMeta{"name": req.Name})
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		pc.Logger.Error(ctx, "ProductController.go", "CreateProduct", "service error", utils.LoggerMeta{"error": err.Error()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	pc.Logger.Info(ctx, "ProductController.go", "CreateProduct", "product created", utils.LoggerMeta{"id": product.ID.String()})
	return ctx.Status(fiber.StatusCreated).JSON(product)
}

// ListMyProducts godoc
// @Summary      List my products
// @Description  List products owned by the authenticated user, newest first
// @Tags         products
// @Produce      json
// @Param        limit   query  int  false  "Page size (20|50|100)"  default(20)
// @Param        offset  query  int  false  "Offset"             default(0)
// @Success      200     {object}  v1.ListProductsResponse
// @Failure      401     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /my-products [get]
func (pc *ProductController) ListMyProducts(ctx fiber.Ctx) error {
	pc.Logger.Debug(ctx, "ProductController.go", "ListMyProducts", "request received", nil)

	var query v1.ListProductsQuery

	if err := ctx.Bind().Query(&query); err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "ListMyProducts", "invalid query parameters", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
	}

	if err := pc.Validator.Struct(query); err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "ListMyProducts", "validation failed", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	if query.Limit == 0 {
		query.Limit = 20
	}

	userID := ctx.Locals(utils.UserContextKey).(uuid.UUID)

	response, err := pc.Service.ListMyProducts(ctx, userID, query.Limit, query.Offset)
	if err != nil {
		pc.Logger.Error(ctx, "ProductController.go", "ListMyProducts", "service error", utils.LoggerMeta{"error": err.Error()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	pc.Logger.Debug(ctx, "ProductController.go", "ListMyProducts", "success", utils.LoggerMeta{"limit": query.Limit, "offset": query.Offset})
	return ctx.JSON(response)
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
	pc.Logger.Debug(ctx, "ProductController.go", "GetProductById", "request received", utils.LoggerMeta{"id": ctx.Params("id")})

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "GetProductById", "invalid product id", utils.LoggerMeta{"id": ctx.Params("id")})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	product, err := pc.Service.GetProductById(ctx, id)
	if err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "GetProductById", "product not found", utils.LoggerMeta{"id": id.String()})
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "product not found",
		})
	}

	pc.Logger.Debug(ctx, "ProductController.go", "GetProductById", "success", utils.LoggerMeta{"id": id.String()})
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
	pc.Logger.Debug(ctx, "ProductController.go", "GetProductByName", "request received", utils.LoggerMeta{"name": name})

	product, err := pc.Service.GetProductByName(ctx, name)
	if err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "GetProductByName", "product not found", utils.LoggerMeta{"name": name})
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "product not found",
		})
	}

	pc.Logger.Debug(ctx, "ProductController.go", "GetProductByName", "success", utils.LoggerMeta{"name": name})
	return ctx.JSON(product)
}

// UpdateProduct godoc
// @Summary      Update a product
// @Description  Update a product's own fields. Categories are managed via link/unlink routes.
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      string              true  "Product ID"
// @Param        product  body      v1.RequestProduct   true  "Product payload"
// @Success      200      {object}  v1.ResponseProduct
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /products/{id} [put]
func (pc *ProductController) UpdateProduct(ctx fiber.Ctx) error {
	pc.Logger.Debug(ctx, "ProductController.go", "UpdateProduct", "request received", utils.LoggerMeta{"id": ctx.Params("id")})

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "UpdateProduct", "invalid product id", utils.LoggerMeta{"id": ctx.Params("id")})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	var req v1.RequestProduct

	if err := ctx.Bind().JSON(&req); err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "UpdateProduct", "invalid request body", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if err := pc.Validator.Struct(req); err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "UpdateProduct", "validation failed", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	product, err := pc.Service.UpdateProduct(ctx, id, req)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateProductName) {
			pc.Logger.Warn(ctx, "ProductController.go", "UpdateProduct", "duplicate product name", utils.LoggerMeta{"id": id.String(), "name": req.Name})
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		pc.Logger.Error(ctx, "ProductController.go", "UpdateProduct", "service error", utils.LoggerMeta{"error": err.Error(), "id": id.String()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	pc.Logger.Info(ctx, "ProductController.go", "UpdateProduct", "product updated", utils.LoggerMeta{"id": id.String()})
	return ctx.JSON(product)
}

// LinkCategory godoc
// @Summary      Link a category to a product
// @Description  Link an existing category to a product. One category per call.
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      string                    true  "Product ID"
// @Param        payload  body      v1.LinkCategoryRequest    true  "Category to link"
// @Success      200      {object}  v1.ResponseProduct
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /products/{id}/categories/link [post]
func (pc *ProductController) LinkCategory(ctx fiber.Ctx) error {
	pc.Logger.Debug(ctx, "ProductController.go", "LinkCategory", "request received", utils.LoggerMeta{"id": ctx.Params("id")})

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "LinkCategory", "invalid product id", utils.LoggerMeta{"id": ctx.Params("id")})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	var req v1.LinkCategoryRequest

	if err := ctx.Bind().JSON(&req); err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "LinkCategory", "invalid request body", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if err := pc.Validator.Struct(req); err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "LinkCategory", "validation failed", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	product, err := pc.Service.LinkCategory(ctx, id, req.CategoryID)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) || errors.Is(err, services.ErrCategoryNotFound) {
			pc.Logger.Warn(ctx, "ProductController.go", "LinkCategory", "not found", utils.LoggerMeta{"error": err.Error()})
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		pc.Logger.Error(ctx, "ProductController.go", "LinkCategory", "service error", utils.LoggerMeta{"error": err.Error(), "product_id": id.String()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	pc.Logger.Info(ctx, "ProductController.go", "LinkCategory", "category linked", utils.LoggerMeta{"product_id": id.String(), "category_id": req.CategoryID.String()})
	return ctx.JSON(product)
}

// UnlinkCategory godoc
// @Summary      Unlink a category from a product
// @Description  Remove a category link from a product. One category per call.
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      string                    true  "Product ID"
// @Param        payload  body      v1.LinkCategoryRequest    true  "Category to unlink"
// @Success      200      {object}  v1.ResponseProduct
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /products/{id}/categories/unlink [post]
func (pc *ProductController) UnlinkCategory(ctx fiber.Ctx) error {
	pc.Logger.Debug(ctx, "ProductController.go", "UnlinkCategory", "request received", utils.LoggerMeta{"id": ctx.Params("id")})

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "UnlinkCategory", "invalid product id", utils.LoggerMeta{"id": ctx.Params("id")})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	var req v1.LinkCategoryRequest

	if err := ctx.Bind().JSON(&req); err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "UnlinkCategory", "invalid request body", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if err := pc.Validator.Struct(req); err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "UnlinkCategory", "validation failed", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	product, err := pc.Service.UnlinkCategory(ctx, id, req.CategoryID)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) || errors.Is(err, services.ErrCategoryNotFound) {
			pc.Logger.Warn(ctx, "ProductController.go", "UnlinkCategory", "not found", utils.LoggerMeta{"error": err.Error()})
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		pc.Logger.Error(ctx, "ProductController.go", "UnlinkCategory", "service error", utils.LoggerMeta{"error": err.Error(), "product_id": id.String()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	pc.Logger.Info(ctx, "ProductController.go", "UnlinkCategory", "category unlinked", utils.LoggerMeta{"product_id": id.String(), "category_id": req.CategoryID.String()})
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
	pc.Logger.Debug(ctx, "ProductController.go", "DeleteProduct", "request received", utils.LoggerMeta{"id": ctx.Params("id")})

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		pc.Logger.Warn(ctx, "ProductController.go", "DeleteProduct", "invalid product id", utils.LoggerMeta{"id": ctx.Params("id")})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id",
		})
	}

	err = pc.Service.DeleteProduct(ctx, id)
	if err != nil {
		pc.Logger.Error(ctx, "ProductController.go", "DeleteProduct", "service error", utils.LoggerMeta{"error": err.Error(), "id": id.String()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	pc.Logger.Info(ctx, "ProductController.go", "DeleteProduct", "product deleted", utils.LoggerMeta{"id": id.String()})
	return ctx.Status(fiber.StatusNoContent).Send(nil)
}
