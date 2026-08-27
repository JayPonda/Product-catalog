package middleware

import (
	"fmt"
	"time"

	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/gofiber/fiber/v3"
)

func RequestLogger(logger *utils.StructuredLogger) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		start := time.Now()

		err := ctx.Next()

		duration := time.Since(start).Milliseconds()
		status := ctx.Response().StatusCode()
		method := ctx.Method()
		path := ctx.Path()
		payloadSize := len(ctx.Response().Body())
		clientIP := ctx.IP()

		logger.Info(ctx, "RequestLogger.go", "Handle", fmt.Sprintf("%s %s", method, path), utils.LoggerMeta{
			"status":        status,
			"duration_ms":   duration,
			"payload_bytes": payloadSize,
			"client_ip":     clientIP,
		})

		return err
	}
}
