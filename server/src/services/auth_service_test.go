package services_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/src/services"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/testdb"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func newAuthService(t *testing.T) *services.AuthService {
	t.Helper()
	db := testdb.OpenSQLite(t)
	repo, err := repositories.InitUserRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatal(err)
	}
	svc, err := services.InitAuthService(db, utils.NewStructuredLogger(), repo, "test-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func registerReq(email string) v1.RegisterRequest {
	return v1.RegisterRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     email,
		Password:  "supersecret1",
	}
}

func TestAuthService_Register_HappyPath_E2E(t *testing.T) {
	svc := newAuthService(t)

	resp, err := svc.Register(registerReq("jane@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.User.Email != "jane@example.com" || resp.User.FirstName != "Jane" {
		t.Errorf("unexpected user response: %+v", resp.User)
	}

	// Persisted user must carry a bcrypt hash, never the raw password.
	user, err := svc.UserManager.GetUserByEmail("jane@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.Password == "supersecret1" {
		t.Error("raw password stored")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("supersecret1")) != nil {
		t.Error("stored hash does not match password")
	}
}

func TestAuthService_Register_DuplicateEmail_E2E(t *testing.T) {
	svc := newAuthService(t)

	if _, err := svc.Register(registerReq("dup@example.com")); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(registerReq("dup@example.com")); !errors.Is(err, services.ErrDuplicateEmail) {
		t.Errorf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestAuthService_Login_Success_E2E(t *testing.T) {
	svc := newAuthService(t)

	reg, err := svc.Register(registerReq("login@example.com"))
	if err != nil {
		t.Fatal(err)
	}

	result, err := svc.Login(v1.LoginRequest{Email: "login@example.com", Password: "supersecret1"})
	if err != nil {
		t.Fatal(err)
	}

	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Fatal("expected both tokens to be issued")
	}
	if result.User.ID.String() != reg.User.ID {
		t.Errorf("wrong user returned: %s vs %s", result.User.ID, reg.User.ID)
	}

	claims, err := utils.ParseAccessToken(result.AccessToken, "test-secret")
	if err != nil {
		t.Fatalf("access token invalid: %v", err)
	}
	if claims.UserID != reg.User.ID {
		t.Errorf("token subject mismatch: %s vs %s", claims.UserID, reg.User.ID)
	}

	hash := utils.HashToken(result.RefreshToken)
	stored, err := svc.UserManager.GetRefreshTokenByHash(hash)
	if err != nil {
		t.Fatalf("refresh token not persisted: %v", err)
	}
	if stored.UserID != result.User.ID {
		t.Error("refresh token linked to wrong user")
	}
	if !stored.ExpiresAt.After(time.Now()) {
		t.Error("refresh token already expired")
	}
}

func TestAuthService_Login_WrongPassword_E2E(t *testing.T) {
	svc := newAuthService(t)
	if _, err := svc.Register(registerReq("pw@example.com")); err != nil {
		t.Fatal(err)
	}

	_, err := svc.Login(v1.LoginRequest{Email: "pw@example.com", Password: "wrongpassword"})
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_UnknownEmail_E2E(t *testing.T) {
	svc := newAuthService(t)

	_, err := svc.Login(v1.LoginRequest{Email: "ghost@example.com", Password: "whatever1"})
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_RefreshToken_ExpiredInvisible_E2E(t *testing.T) {
	svc := newAuthService(t)
	repo := svc.UserManager

	expiredToken := models.RefreshToken{
		UserID:    uuid.New(),
		TokenHash: utils.HashToken("stale-raw-token"),
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	if err := repo.CreateRefreshToken(expiredToken); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.GetRefreshTokenByHash(expiredToken.TokenHash); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expired token visible: %v", err)
	}
}

func TestAuthService_Logout_RevokesAllTokens_E2E(t *testing.T) {
	svc := newAuthService(t)

	reg, err := svc.Register(registerReq("bye@example.com"))
	if err != nil {
		t.Fatal(err)
	}

	login1, err := svc.Login(v1.LoginRequest{Email: "bye@example.com", Password: "supersecret1"})
	if err != nil {
		t.Fatal(err)
	}
	// A second session must also die on logout.
	if _, err := svc.Login(v1.LoginRequest{Email: "bye@example.com", Password: "supersecret1"}); err != nil {
		t.Fatal(err)
	}

	if err := svc.Logout(mustParse(t, reg.User.ID)); err != nil {
		t.Fatal(err)
	}

	hash := utils.HashToken(login1.RefreshToken)
	if _, err := svc.UserManager.GetRefreshTokenByHash(hash); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("token survived logout: %v", err)
	}
}
