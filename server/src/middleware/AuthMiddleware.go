package middleware

import (
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RequireAuth parses the access_token cookie, validates it, and injects the
// user ID into ctx.Locals(utils.UserContextKey). Apply to protected routes.
func RequireAuth(secret string) fiber.Handler {
	l := utils.NewStructuredLogger()
	return func(ctx fiber.Ctx) error {
		tokenString := ctx.Cookies("access_token")
		if tokenString == "" {
			l.Warn("AuthMiddleware.go", "RequireAuth", "missing access token", nil)
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing access token",
			})
		}

		claims, err := utils.ParseAccessToken(tokenString, secret)
		if err != nil {
			l.Error("AuthMiddleware.go", "RequireAuth", "invalid token", utils.LoggerMeta{"error": err.Error()}, "")
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired access token",
			})
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			l.Error("AuthMiddleware.go", "RequireAuth", "invalid user id in token", utils.LoggerMeta{"error": err.Error()}, "")
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid user id in token",
			})
		}

		l.Debug("AuthMiddleware.go", "RequireAuth", "token validated", utils.LoggerMeta{"user_id": userID.String()})
		ctx.Locals(utils.UserContextKey, userID)
		return ctx.Next()
	}
}
