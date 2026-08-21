package routes

import (
	controllersv1 "github.com/JayPonda/Product-catalog/server/src/controllers/v1"
	"github.com/gofiber/fiber/v3"
)

func RegisterV1Routes(app *fiber.App, productController *controllersv1.ProductController, categoryController *controllersv1.CategoryController) {
	v1 := app.Group("/api/v1")

	products := v1.Group("/products")
	products.Get("", productController.ListProducts)
	products.Post("", productController.CreateProduct)
	products.Get("/:id", productController.GetProductById)
	products.Get("/name/:name", productController.GetProductByName)
	products.Put("/:id", productController.UpdateProduct)
	products.Post("/:id/categories/link", productController.LinkCategory)
	products.Post("/:id/categories/unlink", productController.UnlinkCategory)
	products.Delete("/:id", productController.DeleteProduct)

	categories := v1.Group("/categories")
	categories.Get("", categoryController.ListCategories)
	categories.Post("", categoryController.CreateCategory)
	categories.Get("/match", categoryController.MatchCategories)
}
