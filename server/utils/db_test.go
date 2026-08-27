package utils

import (
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3"
)

func TestGetDB_NilContext(t *testing.T) {
	// Reset singleton for clean test
	dbOnce = sync.Once{}
	goquDBInstance = nil

	db := GetDB(nil)
	// Should return nil since not initialized
	if db != nil {
		t.Error("expected nil db when not initialized")
	}
}

func TestGetDB_ContextWithDB(t *testing.T) {
	// Reset singleton for clean test
	dbOnce = sync.Once{}
	goquDBInstance = nil

	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		mockDB := &goqu.Database{}
		c.Locals(DbContextKey, mockDB)
		db := GetDB(c)
		if db != mockDB {
			t.Error("expected GetDB to return db injected via context locals")
		}
		return nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatal(err)
	}
}

func TestGetDB_ContextWithoutDBFallsBackToSingleton(t *testing.T) {
	// Reset singleton for clean test
	dbOnce = sync.Once{}
	goquDBInstance = nil

	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		// Context has no db local => falls back to singleton (nil here)
		db := GetDB(c)
		if db != nil {
			t.Error("expected fallback to nil singleton when no db in locals")
		}
		return nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatal(err)
	}
}
