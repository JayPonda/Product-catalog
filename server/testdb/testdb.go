// Package testdb provides a real, in-memory SQLite database (modernc.org/sqlite,
// pure Go / no CGO) for end-to-end tests. It applies a schema mirrored from the
// production Postgres migrations so repository/service code can be exercised
// against an actual storage engine with zero mocking.
package testdb

import (
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "modernc.org/sqlite"
)

// schemaSQL is the SQLite-mirrored schema for the tables exercised by tests.
// It mirrors the production Postgres migrations closely enough for
// repository/service code to run unchanged. Timestamp columns use DATETIME
// affinity (the SQLite driver only scans those back into time.Time) and
// DEFAULT expressions matching the app's expectation of DB-side timestamps.
const schemaSQL = `
CREATE TABLE IF NOT EXISTS users (
    id          TEXT PRIMARY KEY,
    first_name  TEXT,
    last_name   TEXT,
    email       TEXT,
    password    TEXT,
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    deleted_at  DATETIME
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    token_hash  TEXT NOT NULL,
    expires_at  DATETIME NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    deleted_at  DATETIME
);

CREATE TABLE IF NOT EXISTS orders (
    id          TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    total_bill  REAL NOT NULL,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME,
    deleted_at  DATETIME
);

CREATE TABLE IF NOT EXISTS categories (
    id          TEXT PRIMARY KEY,
    name        TEXT,
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    deleted_at  DATETIME
);

CREATE TABLE IF NOT EXISTS products (
    id             TEXT PRIMARY KEY,
    name           TEXT,
    description    TEXT,
    price          INTEGER,
    stock_quantity INTEGER,
    user_id        TEXT,
    created_at     DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at     DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    deleted_at     DATETIME
);

CREATE TABLE IF NOT EXISTS product_categories (
    id          TEXT PRIMARY KEY,
    product_id  TEXT NOT NULL,
    category_id TEXT NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    deleted_at  DATETIME
);

-- Mirrors migration 20260821031940_uq_products_name_active.sql exactly
-- (SQLite supports partial indexes, same as the Postgres original).
CREATE UNIQUE INDEX IF NOT EXISTS uq_products_name_active
    ON products(name)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_name_active
    ON categories(name)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email_active
    ON users(email)
    WHERE deleted_at IS NULL;
`

var dbCounter uint64

// OpenSQLite spins up a fresh in-memory SQLite database and applies the mirrored
// schema. Each call gets a uniquely named database (cache=shared) so tests stay
// isolated. The connection is closed automatically when the test ends.
func OpenSQLite(t *testing.T) *goqu.Database {
	t.Helper()

	name := fmt.Sprintf("testdb_%d", atomic.AddUint64(&dbCounter, 1))
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("testdb: open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	// Serialize all access through one connection: shared-cache in-memory
	// databases reject concurrent writers (SQLITE_LOCKED/BUSY), and some
	// suites (the dedup job's worker pool) intentionally open several
	// transactions in parallel. Pooling on a single connection keeps those
	// runs deterministic without changing semantics for sequential suites.
	sqlDB.SetMaxOpenConns(1)

	if _, err := sqlDB.Exec(schemaSQL); err != nil {
		t.Fatalf("testdb: apply schema: %v", err)
	}

	return goqu.New("sqlite3", sqlDB)
}

// SeedProduct inserts a single product row with explicit values, giving tests
// deterministic control over created_at for ordering assertions.
func SeedProduct(t *testing.T, db *goqu.Database, id string, name string, createdAt time.Time) {
	t.Helper()
	if _, err := db.Insert("products").Rows(goqu.Record{
		"id":             id,
		"name":           name,
		"description":    "",
		"price":          100,
		"stock_quantity": 0,
		"user_id":        nil,
		"created_at":     createdAt,
		"updated_at":     createdAt,
		"deleted_at":     nil,
	}).Executor().Exec(); err != nil {
		t.Fatalf("testdb: seed product: %v", err)
	}
}

// SeedOrder inserts a single order row directly, bypassing the application's
// insert path (which lives in the seed CLI). Useful for arranging dedup / range
// test fixtures. createdAt is passed as a time.Time so the SQLite driver stores
// and returns it as a time.Time (a formatted string would be returned as a plain
// string that goqu cannot scan into time.Time).
func SeedOrder(t *testing.T, db *goqu.Database, id, customerID string, totalBill float64, createdAt time.Time) {
	t.Helper()
	if _, err := db.Insert("orders").Rows(goqu.Record{
		"id":          id,
		"customer_id": customerID,
		"total_bill":  totalBill,
		"created_at":  createdAt,
		"updated_at":  createdAt,
		"deleted_at":  nil,
	}).Executor().Exec(); err != nil {
		t.Fatalf("testdb: seed order: %v", err)
	}
}
