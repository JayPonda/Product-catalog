package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestStructuredLogger_NoPanicNilCtx(t *testing.T) {
	l := NewStructuredLogger()
	// All level methods must accept a nil fiber.Ctx without panicking.
	l.Debug(nil, "src", "id", "msg", nil)
	l.Info(nil, "src", "id", "msg", LoggerMeta{"k": "v"})
	l.Warn(nil, "src", "id", "msg", nil)
	l.Error(nil, "src", "id", "msg", nil, "trace")
}

func TestGetLoggerFallback(t *testing.T) {
	l := getLogger(nil)
	if l == nil {
		t.Fatal("expected a fallback logger, got nil")
	}
	if _, ok := l.(*StructuredLogger); !ok {
		t.Errorf("expected *StructuredLogger, got %T", l)
	}
}

func TestGetLoggerFromCtx(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(ctx fiber.Ctx) error {
		mockLogger := NewStructuredLogger()
		ctx.Locals(LoggerContextKey, mockLogger)

		l := getLogger(ctx)
		if l != mockLogger {
			t.Error("expected injected logger, got different one")
		}
		return ctx.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test: %v", err)
	}
	res.Body.Close()
}
