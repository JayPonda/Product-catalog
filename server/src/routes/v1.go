package routes

import (
	controllersv1 "github.com/JayPonda/Product-catalog/server/src/controllers/v1"
	"github.com/JayPonda/Product-catalog/server/src/middleware"
	"github.com/gofiber/fiber/v3"
)

func RegisterV1Routes(app *fiber.App, productController *controllersv1.ProductController, categoryController *controllersv1.CategoryController, authController *controllersv1.AuthController, authSecret string) {
	v1 := app.Group("/api/v1")

	products := v1.Group("/products")
	products.Get("", productController.ListProducts)
	products.Get("/:id", productController.GetProductById)
	products.Get("/name/:name", productController.GetProductByName)

	// Mutating routes require a valid access token.
	products.Post("", middleware.RequireAuth(authSecret), productController.CreateProduct)
	products.Put("/:id", middleware.RequireAuth(authSecret), productController.UpdateProduct)
	products.Post("/:id/categories/link", middleware.RequireAuth(authSecret), productController.LinkCategory)
	products.Post("/:id/categories/unlink", middleware.RequireAuth(authSecret), productController.UnlinkCategory)
	products.Delete("/:id", middleware.RequireAuth(authSecret), productController.DeleteProduct)

	categories := v1.Group("/categories")
	categories.Get("", categoryController.ListCategories)
	categories.Get("/match", categoryController.MatchCategories)

	// Mutating routes require a valid access token.
	categories.Post("", middleware.RequireAuth(authSecret), categoryController.CreateCategory)

	auth := v1.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Get("/me", middleware.RequireAuth(authSecret), authController.Me)
	auth.Post("/logout", middleware.RequireAuth(authSecret), authController.Logout)
}
