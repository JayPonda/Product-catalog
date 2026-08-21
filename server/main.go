package main

import (
	// standard library
	"fmt"
	"log"
	"strings"
	"time"

	// third party library
	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	swagger "github.com/gofiber/contrib/v3/swaggo"
	"github.com/joho/godotenv"

	// our own packages
	"github.com/JayPonda/Product-catalog/server/docs"
	controllersv1 "github.com/JayPonda/Product-catalog/server/src/controllers/v1"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	routes "github.com/JayPonda/Product-catalog/server/src/routes"
	"github.com/JayPonda/Product-catalog/server/src/services"
	"github.com/JayPonda/Product-catalog/server/utils"
)

// 1. One common struct with reflection tags managing all configurations
type EnvConfig struct {
	// Web Server Parameters (with default fallbacks)
	AppHost string `env:"HOST" envDefault:"localhost"`
	AppPort string `env:"PORT" envDefault:"8080"`

	// Individual Database Connection Segments parsed natively
	DBHost     string `env:"DB_HOST,required"`
	DBPort     string `env:"DB_PORT,required"`
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBName     string `env:"DB_NAME,required"`
	DBSSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`

	// Pool Settings (Parsed into integers and time.Duration automatically)
	DBMaxOpen     int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	DBMaxIdle     int           `env:"DB_MAX_IDLE_CONNS" envDefault:"25"`
	DBMaxLifetime time.Duration `env:"DB_MAX_LIFETIME" envDefault:"5m"`
	DBMaxIdleTime time.Duration `env:"DB_MAX_IDLE_TIME" envDefault:"2m"`

	// CORS
	AllowedOrigins string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://localhost:5173"`
}

// 2. Local getter implementations
func (e *EnvConfig) getHost() string { return e.AppHost }
func (e *EnvConfig) getPort() string { return e.AppPort }

// 3. Fulfill your read-only DBConfigProvider contract requirements
func (e *EnvConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		e.DBHost, e.DBPort, e.DBUser, e.DBPassword, e.DBName, e.DBSSLMode)
}
func (e *EnvConfig) GetMaxOpenConns() int          { return e.DBMaxOpen }
func (e *EnvConfig) GetMaxIdleConns() int          { return e.DBMaxIdle }
func (e *EnvConfig) GetMaxLifetime() time.Duration { return e.DBMaxLifetime }
func (e *EnvConfig) GetMaxIdleTime() time.Duration { return e.DBMaxIdleTime }

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from system environment")
	}

	// 4. Initialize and parse using the reflection engine wrapper
	cfg := &EnvConfig{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Critical: Failed to parse configuration tags: %v", err)
	}

	// Initialize your framework engine
	app := fiber.New()

	// CORS middleware
	origins := strings.Split(cfg.AllowedOrigins, ",")
	app.Use(cors.New(cors.Config{
		AllowOrigins: origins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// Register structured logging layer middleware properties
	appLogger := utils.NewStructuredLogger()
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

	// Initialize services
	productService, err := services.InitProductService(orm, appLogger, productRepo, categoryRepo, productCategoryRepo)
	if err != nil {
		log.Fatalf("Failed to initialize product service: %v", err)
	}

	categoryService, err := services.InitCategoryService(appLogger, categoryRepo)
	if err != nil {
		log.Fatalf("Failed to initialize category service: %v", err)
	}

	// Initialize controllers
	productController := controllersv1.NewProductController(productService)
	categoryController := controllersv1.NewCategoryController(categoryService)

	// Register routes
	routes.RegisterV1Routes(app, productController, categoryController)

	// Swagger docs
	docs.SwaggerInfo.BasePath = "/api/v1"
	app.Get("/docs/*", swagger.HandlerDefault)

	// Health check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("OK")
	})

	// Boot up application listener bindings dynamically using structural values
	log.Fatal(app.Listen(fmt.Sprintf("%s:%s", cfg.getHost(), cfg.getPort())))
}





