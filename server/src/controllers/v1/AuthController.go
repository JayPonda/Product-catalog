package controllersv1

import (
	"errors"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/services"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const (
	accessCookieName  = "access_token"
	refreshCookieName = "refresh_token"
)

type AuthController struct {
	Service       *services.AuthService
	Validator     *validator.Validate
	SecureCookies bool
	Logger        *utils.StructuredLogger
}

func NewAuthController(service *services.AuthService, secureCookies bool, logger *utils.StructuredLogger) *AuthController {
	return &AuthController{
		Service:       service,
		Validator:     utils.NewValidator(),
		SecureCookies: secureCookies,
		Logger:        logger,
	}
}

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload body      v1.RegisterRequest true "Registration payload"
// @Success      201     {object}  v1.AuthResponse
// @Failure      400     {object}  map[string]string
// @Failure      409     {object}  map[string]string
// @Router       /auth/register [post]
func (ac *AuthController) Register(ctx fiber.Ctx) error {
	ac.Logger.Debug("AuthController.go", "Register", "request received", nil)

	var req v1.RegisterRequest

	if err := ctx.Bind().JSON(&req); err != nil {
		ac.Logger.Warn("AuthController.go", "Register", "invalid request body", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if err := ac.Validator.Struct(req); err != nil {
		ac.Logger.Warn("AuthController.go", "Register", "validation failed", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	response, err := ac.Service.Register(req)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateEmail) {
			ac.Logger.Warn("AuthController.go", "Register", "duplicate email", utils.LoggerMeta{"email": req.Email})
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		ac.Logger.Error("AuthController.go", "Register", "service error", utils.LoggerMeta{"error": err.Error()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	ac.Logger.Info("AuthController.go", "Register", "user registered", utils.LoggerMeta{"user_id": response.User.ID})
	return ctx.Status(fiber.StatusCreated).JSON(response)
}

// Login godoc
// @Summary      Login and receive auth cookies
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload body      v1.LoginRequest true "Login payload"
// @Success      200     {object}  v1.AuthResponse
// @Failure      400     {object}  map[string]string
// @Failure      401     {object}  map[string]string
// @Router       /auth/login [post]
func (ac *AuthController) Login(ctx fiber.Ctx) error {
	ac.Logger.Debug("AuthController.go", "Login", "request received", nil)

	var req v1.LoginRequest

	if err := ctx.Bind().JSON(&req); err != nil {
		ac.Logger.Warn("AuthController.go", "Login", "invalid request body", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if err := ac.Validator.Struct(req); err != nil {
		ac.Logger.Warn("AuthController.go", "Login", "validation failed", utils.LoggerMeta{"error": err.Error()})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	result, err := ac.Service.Login(req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			ac.Logger.Warn("AuthController.go", "Login", "invalid credentials", utils.LoggerMeta{"email": req.Email})
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		ac.Logger.Error("AuthController.go", "Login", "service error", utils.LoggerMeta{"error": err.Error()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	ac.setAuthCookies(ctx, result.AccessToken, result.RefreshToken)

	ac.Logger.Info("AuthController.go", "Login", "login successful", utils.LoggerMeta{"user_id": result.User.ID.String()})
	return ctx.Status(fiber.StatusOK).JSON(v1.AuthResponse{
		User: v1.ToUserResponse(result.User),
	})
}

// Logout godoc
// @Summary      Logout (revoke refresh token)
// @Tags         auth
// @Security     CookieAuth
// @Success      204
// @Failure      401 {object} map[string]string
// @Router       /auth/logout [post]
func (ac *AuthController) Logout(ctx fiber.Ctx) error {
	ac.Logger.Debug("AuthController.go", "Logout", "request received", nil)

	userID, ok := ctx.Locals(utils.UserContextKey).(uuid.UUID)
	if !ok {
		ac.Logger.Warn("AuthController.go", "Logout", "unauthenticated", nil)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthenticated",
		})
	}

	if err := ac.Service.Logout(userID); err != nil {
		ac.Logger.Error("AuthController.go", "Logout", "service error", utils.LoggerMeta{"error": err.Error(), "user_id": userID.String()}, "")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	ac.Logger.Info("AuthController.go", "Logout", "logout successful", utils.LoggerMeta{"user_id": userID.String()})
	ac.clearAuthCookies(ctx)
	return ctx.Status(fiber.StatusNoContent).Send(nil)
}

// Me godoc
// @Summary      Get the currently authenticated user
// @Tags         auth
// @Security     CookieAuth
// @Success      200 {object} v1.AuthResponse
// @Failure      401 {object} map[string]string
// @Router       /auth/me [get]
func (ac *AuthController) Me(ctx fiber.Ctx) error {
	ac.Logger.Debug("AuthController.go", "Me", "request received", nil)

	userID, ok := ctx.Locals(utils.UserContextKey).(uuid.UUID)
	if !ok {
		ac.Logger.Warn("AuthController.go", "Me", "unauthenticated", nil)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthenticated",
		})
	}

	user, err := ac.Service.UserManager.GetUserById(userID)
	if err != nil {
		ac.Logger.Warn("AuthController.go", "Me", "user not found", utils.LoggerMeta{"user_id": userID.String()})
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	ac.Logger.Debug("AuthController.go", "Me", "success", utils.LoggerMeta{"user_id": userID.String()})
	return ctx.JSON(v1.AuthResponse{
		User: v1.ToUserResponse(user),
	})
}

func (ac *AuthController) setAuthCookies(ctx fiber.Ctx, accessToken, refreshToken string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     accessCookieName,
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   ac.SecureCookies,
		SameSite: "Lax",
		Path:     "/",
	})
	ctx.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   ac.SecureCookies,
		SameSite: "Lax",
		Path:     "/",
	})
}

func (ac *AuthController) clearAuthCookies(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     accessCookieName,
		Value:    "",
		HTTPOnly: true,
		Secure:   ac.SecureCookies,
		SameSite: "Lax",
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})
	ctx.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		HTTPOnly: true,
		Secure:   ac.SecureCookies,
		SameSite: "Lax",
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})
}
