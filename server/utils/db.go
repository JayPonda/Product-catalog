package utils

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3"
	_ "github.com/lib/pq"
)

// 1. The minimal interface that demands configuration values from the outside
type DBConfigProvider interface {
	GetDSN() string
	GetDialect() string
	GetMaxOpenConns() int
	GetMaxIdleConns() int
	GetMaxLifetime() time.Duration
	GetMaxIdleTime() time.Duration
}

const DbContextKey = "app_db"

var (
	goquDBInstance *goqu.Database
	dbOnce         sync.Once
)

// 2. InitDB now demands the interface as an argument instead of a raw DSN string
func InitDB(cfg DBConfigProvider) *goqu.Database {
	dbOnce.Do(func() {
		// 1. Open the connection pool using the DSN from the outside config
		db, err := sql.Open(cfg.GetDialect(), cfg.GetDSN())
		if err != nil {
			log.Fatalf("Critical: Error opening database stream: %v", err)
		}

		// 2. Configure pool resource allocations using values from the outside config
		db.SetMaxOpenConns(cfg.GetMaxOpenConns())
		db.SetMaxIdleConns(cfg.GetMaxIdleConns())
		db.SetConnMaxLifetime(cfg.GetMaxLifetime())
		db.SetConnMaxIdleTime(cfg.GetMaxIdleTime())

		// 3. Verify physical network connectivity immediately
		if err := db.Ping(); err != nil {
			log.Fatalf("Critical: Database is unreachable: %v", err)
		}

		// 4. Wrap with goqu dialect and store into singleton state
		dialect := goqu.Dialect(cfg.GetDialect())
		goquDBInstance = dialect.DB(db)

		log.Println("Database connection established successfully")
	})

	return goquDBInstance
}

// GetDB extracts the database pointer from Fiber's request context
func GetDB(ctx fiber.Ctx) *goqu.Database {
	if ctx != nil {
		if db, ok := ctx.Locals(DbContextKey).(*goqu.Database); ok {
			return db
		}
	}
	return goquDBInstance
}
