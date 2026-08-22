package main

import (
	// standard library
	"fmt"
	"log"
	"time"

	// third party library
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	// our own packages
	cmd "github.com/JayPonda/Product-catalog/server/commands/cmd"
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

	// SQL dialect for migrations (postgres, mysql, sqlite3, sqlserver, ...)
	DBDialect string `env:"DB_DIALECT" envDefault:"postgres"`

	// Pool Settings (Parsed into integers and time.Duration automatically)
	DBMaxOpen     int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	DBMaxIdle     int           `env:"DB_MAX_IDLE_CONNS" envDefault:"25"`
	DBMaxLifetime time.Duration `env:"DB_MAX_LIFETIME" envDefault:"5m"`
	DBMaxIdleTime time.Duration `env:"DB_MAX_IDLE_TIME" envDefault:"2m"`

	// CORS
	AllowedOrigins string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://localhost:5173"`

	// Runtime environment (local|prod). Controls cookie Secure flag.
	AppEnv string `env:"APP_ENV" envDefault:"local"`

	// Auth / JWT
	JWTSecret       string        `env:"JWT_SECRET,required"`
	AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"15m"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL" envDefault:"168h"`
}

// 2. Local getter implementations
func (e *EnvConfig) GetHost() string           { return e.AppHost }
func (e *EnvConfig) GetPort() string           { return e.AppPort }
func (e *EnvConfig) GetAllowedOrigins() string { return e.AllowedOrigins }

// GetAppEnv returns the runtime environment; used to decide cookie Secure flag.
func (e *EnvConfig) GetAppEnv() string { return e.AppEnv }

// Auth / JWT getters
func (e *EnvConfig) GetJWTSecret() string          { return e.JWTSecret }
func (e *EnvConfig) GetAccessTokenTTL() time.Duration  { return e.AccessTokenTTL }
func (e *EnvConfig) GetRefreshTokenTTL() time.Duration { return e.RefreshTokenTTL }

// 3. Fulfill your read-only DBConfigProvider contract requirements
func (e *EnvConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		e.DBHost, e.DBPort, e.DBUser, e.DBPassword, e.DBName, e.DBSSLMode)
}
func (e *EnvConfig) GetDialect() string { return e.DBDialect }
func (e *EnvConfig) GetMaxOpenConns() int          { return e.DBMaxOpen }
func (e *EnvConfig) GetMaxIdleConns() int          { return e.DBMaxIdle }
func (e *EnvConfig) GetMaxLifetime() time.Duration { return e.DBMaxLifetime }
func (e *EnvConfig) GetMaxIdleTime() time.Duration { return e.DBMaxIdleTime }

func main() {
	// Load environment variables from .env (if present), then parse config.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from system environment")
	}

	cfg := &EnvConfig{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Critical: Failed to parse configuration tags: %v", err)
	}

	// Initialize the shared structured logger.
	appLogger := utils.NewStructuredLogger()

	// Dispatch to the CLI: `server` starts the HTTP server, `migrate` manages
	// database migrations. Add new commands in commands/cmd (see root.go).
	cmd.Execute(cfg, appLogger)
}
