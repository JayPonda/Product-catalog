package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const CorrelationIDKey = "correlation_id"

func CorrelationID() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		id := ctx.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}

		ctx.Locals(CorrelationIDKey, id)
		ctx.Set("X-Request-ID", id)

		return ctx.Next()
	}
}

func GetCorrelationID(ctx fiber.Ctx) string {
	if ctx != nil {
		if id, ok := ctx.Locals(CorrelationIDKey).(string); ok {
			return id
		}
	}
	return ""
}
