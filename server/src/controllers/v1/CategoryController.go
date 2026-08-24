package controllersv1

import (
	"errors"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/services"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type CategoryController struct {
	Service   *services.CategoryService
	Validator *validator.Validate
}

func NewCategoryController(service *services.CategoryService) *CategoryController {
	return &CategoryController{
		Service:   service,
		Validator: validator.New(),
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
	var query v1.MatchCategoriesQuery

	if err := ctx.Bind().Query(&query); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
	}

	if err := cc.Validator.Struct(query); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	if query.Limit == 0 {
		query.Limit = 10
	}

	categories, err := cc.Service.MatchCategories(query.Name, query.Limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

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
	var query v1.ListCategoriesQuery

	if err := ctx.Bind().Query(&query); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
	}

	if err := cc.Validator.Struct(query); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	if query.Limit == 0 {
		query.Limit = 20
	}

	response, err := cc.Service.ListCategories(query.Limit, query.Offset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

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
	var req v1.RequestCategory

	if err := ctx.Bind().JSON(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if err := cc.Validator.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	category, err := cc.Service.CreateCategory(models.Category{Name: req.Name})
	if err != nil {
		if errors.Is(err, services.ErrEmptyCategoryName) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if errors.Is(err, services.ErrDuplicateCategoryName) {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(category)
}
