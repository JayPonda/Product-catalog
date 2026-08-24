/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"strings"

	// third party library
	swagger "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/spf13/cobra"

	// our own packages
	"github.com/JayPonda/Product-catalog/server/docs"
	controllersv1 "github.com/JayPonda/Product-catalog/server/src/controllers/v1"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	routes "github.com/JayPonda/Product-catalog/server/src/routes"
	"github.com/JayPonda/Product-catalog/server/src/services"
	"github.com/JayPonda/Product-catalog/server/utils"
)

// serverCmd starts the HTTP server and wires up all dependencies.
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Product Catalog HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := appConfig
		appLogger := appLogger

		// Initialize your framework engine
		app := fiber.New()

		// CORS middleware
		origins := strings.Split(cfg.GetAllowedOrigins(), ",")
		app.Use(cors.New(cors.Config{
			AllowOrigins:     origins,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			AllowCredentials: true,
		}))

		// Register structured logging layer middleware properties
		app.Use(func(ctx fiber.Ctx) error {
			ctx.Locals(utils.LoggerContextKey, appLogger)
			return ctx.Next()
		})

		// Initialize the connection pool singleton by passing your unified configuration structure
		orm := utils.InitDB(cfg)

		// Inject your pool instance safely into every incoming request scope
		app.Use(func(ctx fiber.Ctx) error {
			ctx.Locals(utils.DbContextKey, orm)
			return ctx.Next()
		})

		// Initialize repositories
		productRepo, err := repositories.InitProductRepository(orm, appLogger)
		if err != nil {
			log.Fatalf("Failed to initialize product repository: %v", err)
		}

		categoryRepo, err := repositories.InitCategoryRepository(orm, appLogger)
		if err != nil {
			log.Fatalf("Failed to initialize category repository: %v", err)
		}

		productCategoryRepo, err := repositories.InitProductCategoryRepository(orm, appLogger)
		if err != nil {
			log.Fatalf("Failed to initialize product category repository: %v", err)
		}

		userRepo, err := repositories.InitUserRepository(orm, appLogger)
		if err != nil {
			log.Fatalf("Failed to initialize user repository: %v", err)
		}

		// Initialize services
		productService, err := services.InitProductService(orm, appLogger, productRepo, categoryRepo, productCategoryRepo)
		if err != nil {
			log.Fatalf("Failed to initialize product service: %v", err)
		}

		categoryService, err := services.InitCategoryService(appLogger, categoryRepo)
		if err != nil {
			log.Fatalf("Failed to initialize category service: %v", err)
		}

		authService, err := services.InitAuthService(
			orm,
			appLogger,
			userRepo,
			cfg.GetJWTSecret(),
			cfg.GetAccessTokenTTL(),
			cfg.GetRefreshTokenTTL(),
		)
		if err != nil {
			log.Fatalf("Failed to initialize auth service: %v", err)
		}

		// Initialize controllers
		productController := controllersv1.NewProductController(productService)
		categoryController := controllersv1.NewCategoryController(categoryService)
		authController := controllersv1.NewAuthController(authService, cfg.GetAppEnv() != "local")

		// Register routes
		routes.RegisterV1Routes(app, productController, categoryController, authController, cfg.GetJWTSecret())

		// Swagger docs
		docs.SwaggerInfo.BasePath = "/api/v1"
		app.Get("/docs/*", swagger.HandlerDefault)

		// Health check
		app.Get("/health", func(c fiber.Ctx) error {
			return c.Status(fiber.StatusOK).SendString("OK")
		})

		// Boot up application listener bindings dynamically using structural values
		return app.Listen(fmt.Sprintf("%s:%s", cfg.GetHost(), cfg.GetPort()))
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
