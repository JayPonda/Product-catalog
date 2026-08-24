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
}

func NewAuthController(service *services.AuthService, secureCookies bool) *AuthController {
	return &AuthController{
		Service:       service,
		Validator:     utils.NewValidator(),
		SecureCookies: secureCookies,
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
	var req v1.RegisterRequest

	if err := ctx.Bind().JSON(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if err := ac.Validator.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	response, err := ac.Service.Register(req)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateEmail) {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

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
	var req v1.LoginRequest

	if err := ctx.Bind().JSON(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if err := ac.Validator.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	result, err := ac.Service.Login(req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	ac.setAuthCookies(ctx, result.AccessToken, result.RefreshToken)

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
	userID, ok := ctx.Locals(utils.UserContextKey).(uuid.UUID)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthenticated",
		})
	}

	if err := ac.Service.Logout(userID); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

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
	userID, ok := ctx.Locals(utils.UserContextKey).(uuid.UUID)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthenticated",
		})
	}

	user, err := ac.Service.UserManager.GetUserById(userID)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not found",
		})
	}

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
