package middleware

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestCorrelationID_GeneratesNewID(t *testing.T) {
	app := fiber.New()
	app.Use(CorrelationID())
	app.Get("/test", func(c fiber.Ctx) error {
		id := c.Locals(utils.CorrelationIDKey).(string)
		return c.JSON(fiber.Map{"id": id})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	res, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	body, _ := io.ReadAll(res.Body)
	var resp map[string]string
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["id"] == "" {
		t.Error("expected correlation ID to be generated")
	}
	if res.Header.Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header to be set")
	}
	if resp["id"] != res.Header.Get("X-Request-ID") {
		t.Errorf("expected header and local to match: %s vs %s", resp["id"], res.Header.Get("X-Request-ID"))
	}
}

func TestCorrelationID_PropagatesExistingID(t *testing.T) {
	app := fiber.New()
	app.Use(CorrelationID())
	app.Get("/test", func(c fiber.Ctx) error {
		id := c.Locals(utils.CorrelationIDKey).(string)
		return c.JSON(fiber.Map{"id": id})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "my-custom-id-123")
	res, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	body, _ := io.ReadAll(res.Body)
	var resp map[string]string
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["id"] != "my-custom-id-123" {
		t.Errorf("expected propagated ID, got %s", resp["id"])
	}
	if res.Header.Get("X-Request-ID") != "my-custom-id-123" {
		t.Errorf("expected header to be propagated, got %s", res.Header.Get("X-Request-ID"))
	}
}

func TestRequestLogger_LogsRequest(t *testing.T) {
	app := fiber.New()
	logger := utils.NewStructuredLogger()
	app.Use(RequestLogger(logger))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	res, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
}

func TestRequireAuth_MissingToken(t *testing.T) {
	app := fiber.New()
	logger := utils.NewStructuredLogger()
	app.Use(RequireAuth("test-secret", logger))
	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	res, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", res.StatusCode)
	}

	body, _ := io.ReadAll(res.Body)
	var resp map[string]string
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["error"] != "missing access token" {
		t.Errorf("expected 'missing access token', got %s", resp["error"])
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	app := fiber.New()
	logger := utils.NewStructuredLogger()
	app.Use(RequireAuth("test-secret", logger))
	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Cookie", "access_token=invalid.jwt.token")
	res, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", res.StatusCode)
	}
}

func TestRequireAuth_WrongSecret(t *testing.T) {
	app := fiber.New()
	logger := utils.NewStructuredLogger()
	app.Use(RequireAuth("correct-secret", logger))
	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	token, _ := utils.GenerateAccessToken(uuid.New().String(), "wrong-secret", time.Hour)
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Cookie", "access_token="+token)
	res, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", res.StatusCode)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	app := fiber.New()
	logger := utils.NewStructuredLogger()
	app.Use(RequireAuth("test-secret", logger))
	app.Get("/protected", func(c fiber.Ctx) error {
		userID := c.Locals(utils.UserContextKey).(uuid.UUID)
		return c.JSON(fiber.Map{"user_id": userID.String()})
	})

	userID := uuid.New()
	token, _ := utils.GenerateAccessToken(userID.String(), "test-secret", time.Hour)
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Cookie", "access_token="+token)
	res, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}

	body, _ := io.ReadAll(res.Body)
	var resp map[string]string
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["user_id"] != userID.String() {
		t.Errorf("expected user_id %s, got %s", userID.String(), resp["user_id"])
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	app := fiber.New()
	logger := utils.NewStructuredLogger()
	app.Use(RequireAuth("test-secret", logger))
	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// Create a token with 0 TTL (already expired)
	token, _ := utils.GenerateAccessToken(uuid.New().String(), "test-secret", -time.Hour)
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Cookie", "access_token="+token)
	res, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", res.StatusCode)
	}
}

func TestRequireAuth_InvalidSigningMethod(t *testing.T) {
	app := fiber.New()
	logger := utils.NewStructuredLogger()
	app.Use(RequireAuth("test-secret", logger))
	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// Create a token with none signing method
	claims := jwt.MapClaims{
		"sub": uuid.New().String(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Cookie", "access_token="+tokenString)
	res, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 for none signing, got %d", res.StatusCode)
	}
}

func TestRequireAuth_InvalidClaims(t *testing.T) {
	app := fiber.New()
	logger := utils.NewStructuredLogger()
	app.Use(RequireAuth("test-secret", logger))
	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// Create a token with non-MapClaims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Subject: uuid.New().String(),
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Cookie", "access_token="+tokenString)
	_, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	// This should still work because jwt.ParseWithClaims returns MapClaims
	// But let's test with a token that has no sub
	token2 := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString2, _ := token2.SignedString([]byte("test-secret"))

	req2 := httptest.NewRequest("GET", "/protected", nil)
	req2.Header.Set("Cookie", "access_token="+tokenString2)
	res2, err := app.Test(req2, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	if res2.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 for missing sub, got %d", res2.StatusCode)
	}
}

func TestRequireAuth_InvalidUUIDInSub(t *testing.T) {
	app := fiber.New()
	logger := utils.NewStructuredLogger()
	app.Use(RequireAuth("test-secret", logger))
	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "not-a-valid-uuid",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Cookie", "access_token="+tokenString)
	res, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 for invalid UUID, got %d", res.StatusCode)
	}
}
