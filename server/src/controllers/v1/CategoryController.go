package controllersv1

import (
	"github.com/JayPonda/Product-catalog/server/src/services"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/gofiber/fiber/v3"
	"github.com/go-playground/validator/v10"
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
// @Param        name   query  string  false  "Name prefix"          default(sale)
// @Param        limit  query  int     false  "Max results (1-100)"  default(10)
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

	return ctx.JSON(v1.MatchCategoriesResponse{Categories: categories})
}
