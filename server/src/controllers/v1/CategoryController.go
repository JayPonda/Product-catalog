package controllersv1

import (
	"errors"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/services"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type CategoryController struct {
	Service   *services.CategoryService
	Validator *validator.Validate
	Logger    *utils.StructuredLogger
}

func NewCategoryController(service *services.CategoryService, logger *utils.StructuredLogger) *CategoryController {
	return &CategoryController{
		Service:   service,
		Validator: validator.New(),
		Logger:    logger,
	}
}

// MatchCategories godoc
// @Summary      Match categories by name prefix
// @Description  List active categories whose name starts with the given prefix
// @Tags         categories
// @Produce      json
// @Param        name   query  string  false  "Name prefix"        default(sale)
// @Param        limit  query  int     false  "Max results (5|10|20)"  default(10)
// @Success      200    {object}  v1.MatchCategoriesResponse
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /categories/match [get]
func (cc *CategoryController) MatchCategories(ctx fiber.Ctx) error {
	cc.Logger.Debug(ctx, "CategoryController.go", "MatchCategories", "request received", nil)

	var query v1.MatchCategoriesQuery

	if err := ctx.Bind().Query(&query); err != nil {
		cc.Logger.Warn(ctx, "CategoryController.go", "MatchCategories", "invalid query parameters", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
	}

	if err := cc.Validator.Struct(query); err != nil {
		cc.Logger.Warn(ctx, "CategoryController.go", "MatchCategories", "validation failed", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	if query.Limit == 0 {
		query.Limit = 10
	}

	categories, err := cc.Service.MatchCategories(ctx, query.Name, query.Limit)
	if err != nil {
		cc.Logger.Error(ctx, "CategoryController.go", "MatchCategories", "service error", utils.LoggerMeta{"error": err.Error()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	cc.Logger.Debug(ctx, "CategoryController.go", "MatchCategories", "success", utils.LoggerMeta{"name": query.Name, "limit": query.Limit})
	return ctx.Status(fiber.StatusOK).JSON(v1.MatchCategoriesResponse{Categories: categories})
}

// ListCategories godoc
// @Summary      List categories
// @Description  List categories with pagination, alphabetical by name
// @Tags         categories
// @Produce      json
// @Param        limit   query  int  false  "Page size (20|50|100)"  default(20)
// @Param        offset  query  int  false  "Offset"             default(0)
// @Success      200     {object}  v1.ListCategoriesResponse
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /categories [get]
func (cc *CategoryController) ListCategories(ctx fiber.Ctx) error {
	cc.Logger.Debug(ctx, "CategoryController.go", "ListCategories", "request received", nil)

	var query v1.ListCategoriesQuery

	if err := ctx.Bind().Query(&query); err != nil {
		cc.Logger.Warn(ctx, "CategoryController.go", "ListCategories", "invalid query parameters", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
	}

	if err := cc.Validator.Struct(query); err != nil {
		cc.Logger.Warn(ctx, "CategoryController.go", "ListCategories", "validation failed", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	if query.Limit == 0 {
		query.Limit = 20
	}

	response, err := cc.Service.ListCategories(ctx, query.Limit, query.Offset)
	if err != nil {
		cc.Logger.Error(ctx, "CategoryController.go", "ListCategories", "service error", utils.LoggerMeta{"error": err.Error()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	cc.Logger.Debug(ctx, "CategoryController.go", "ListCategories", "success", utils.LoggerMeta{"limit": query.Limit, "offset": query.Offset})
	return ctx.JSON(response)
}

// CreateCategory godoc
// @Summary      Create a category
// @Description  Create a new category. This is the only way categories are created.
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category  body      v1.RequestCategory  true  "Category payload"
// @Success      201       {object}  models.Category
// @Failure      400       {object}  map[string]string
// @Failure      409       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /categories [post]
func (cc *CategoryController) CreateCategory(ctx fiber.Ctx) error {
	cc.Logger.Debug(ctx, "CategoryController.go", "CreateCategory", "request received", nil)

	var req v1.RequestCategory

	if err := ctx.Bind().JSON(&req); err != nil {
		cc.Logger.Warn(ctx, "CategoryController.go", "CreateCategory", "invalid request body", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if err := cc.Validator.Struct(req); err != nil {
		cc.Logger.Warn(ctx, "CategoryController.go", "CreateCategory", "validation failed", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	category, err := cc.Service.CreateCategory(ctx, models.Category{Name: req.Name})
	if err != nil {
		if errors.Is(err, services.ErrEmptyCategoryName) {
			cc.Logger.Warn(ctx, "CategoryController.go", "CreateCategory", "empty category name", nil)
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if errors.Is(err, services.ErrDuplicateCategoryName) {
			cc.Logger.Warn(ctx, "CategoryController.go", "CreateCategory", "duplicate category name", utils.LoggerMeta{"name": req.Name})
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		cc.Logger.Error(ctx, "CategoryController.go", "CreateCategory", "service error", utils.LoggerMeta{"error": err.Error()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	cc.Logger.Info(ctx, "CategoryController.go", "CreateCategory", "category created", utils.LoggerMeta{"id": category.ID.String(), "name": category.Name})
	return ctx.Status(fiber.StatusCreated).JSON(category)
}
