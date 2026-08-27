package middleware

import (
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func CorrelationID() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		id := ctx.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}

		ctx.Locals(utils.CorrelationIDKey, id)
		ctx.Set("X-Request-ID", id)

		return ctx.Next()
	}
}
