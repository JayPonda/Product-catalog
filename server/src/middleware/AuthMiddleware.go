package middleware

import (
	"strings"

	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func RequireAuth(secret string, logger *utils.StructuredLogger) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		tokenString := ctx.Cookies("access_token")

		if tokenString == "" {
			logger.Warn(ctx, "AuthMiddleware.go", "RequireAuth", "missing access token", nil)
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing access token"})
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			logger.Error(ctx, "AuthMiddleware.go", "RequireAuth", "invalid token", utils.LoggerMeta{"error": err.Error()}, "")
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token claims"})
		}

		sub, ok := claims["sub"].(string)
		if !ok || strings.TrimSpace(sub) == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing subject in token"})
		}

		userID, err := uuid.Parse(sub)
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user id in token"})
		}

		ctx.Locals(utils.UserContextKey, userID)

		logger.Debug(ctx, "AuthMiddleware.go", "RequireAuth", "user authenticated", utils.LoggerMeta{"user_id": userID.String()})

		return ctx.Next()
	}
}
