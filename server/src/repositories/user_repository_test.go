package repositories_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/testdb"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/google/uuid"
)

func newUserRepo(t *testing.T) *repositories.UserRepository {
	t.Helper()
	db := testdb.OpenSQLite(t)
	repo, err := repositories.InitUserRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatalf("init user repo: %v", err)
	}
	return repo
}

func mkUser(name string) models.User {
	return models.User{
		FirstName: name,
		LastName:  "Tester",
		Email:     name + "@example.com",
		Password:  "hashed-password",
	}
}

func TestUserRepo_CreateAndGet_E2E(t *testing.T) {
	repo := newUserRepo(t)

	created, err := repo.CreateUser(nil, mkUser("alice"))
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Error("expected generated id")
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected DB-side created_at default to populate")
	}

	byEmail, err := repo.GetUserByEmail(nil, "alice@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if byEmail.ID != created.ID || byEmail.FirstName != "alice" {
		t.Errorf("unexpected user: %+v", byEmail)
	}
	if byEmail.Password != "hashed-password" {
		t.Error("expected password hash to be selected")
	}

	byID, err := repo.GetUserById(nil, created.ID)
	if err != nil {
		t.Fatalf("GetUserById: %v", err)
	}
	if byID.ID != created.ID {
		t.Errorf("id mismatch: %s vs %s", byID.ID, created.ID)
	}
}

func TestUserRepo_GetUserByEmail_NotFound_E2E(t *testing.T) {
	repo := newUserRepo(t)
	if _, err := repo.GetUserByEmail(nil, "ghost@example.com"); err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}

func TestUserRepo_SoftDelete_HidesUser_E2E(t *testing.T) {
	repo := newUserRepo(t)
	created, err := repo.CreateUser(nil, mkUser("bob"))
	if err != nil {
		t.Fatal(err)
	}

	if err := repo.SoftDeleteUser(nil, created.ID); err != nil {
		t.Fatalf("SoftDeleteUser: %v", err)
	}
	if _, err := repo.GetUserById(nil, created.ID); err != sql.ErrNoRows {
		t.Errorf("soft-deleted user should be invisible, got %v", err)
	}
}

func TestUserRepo_RefreshTokenLifecycle_E2E(t *testing.T) {
	repo := newUserRepo(t)
	user, err := repo.CreateUser(nil, mkUser("carol"))
	if err != nil {
		t.Fatal(err)
	}

	hash := "abc123hash"
	future := time.Now().Add(24 * time.Hour).UTC()
	if err := repo.CreateRefreshToken(nil, models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: future,
	}); err != nil {
		t.Fatalf("CreateRefreshToken: %v", err)
	}

	got, err := repo.GetRefreshTokenByHash(nil, hash)
	if err != nil {
		t.Fatalf("GetRefreshTokenByHash: %v", err)
	}
	if got.UserID != user.ID || got.TokenHash != hash {
		t.Errorf("unexpected token: %+v", got)
	}
	if got.ExpiresAt.Before(time.Now()) {
		t.Error("expires_at should be in the future")
	}
	if got.CreatedAt.IsZero() {
		t.Error("expected DB-side created_at default to populate")
	}

	// Expired tokens must not be returned.
	expired := "expiredhash"
	if err := repo.CreateRefreshToken(nil, models.RefreshToken{
		UserID:    user.ID,
		TokenHash: expired,
		ExpiresAt: time.Now().Add(-time.Hour).UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetRefreshTokenByHash(nil, expired); err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows for expired token, got %v", err)
	}

	// Revoke-all for the user.
	if err := repo.DeleteRefreshTokensByUser(nil, user.ID); err != nil {
		t.Fatalf("DeleteRefreshTokensByUser: %v", err)
	}
	if _, err := repo.GetRefreshTokenByHash(nil, hash); err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows after revoke-all, got %v", err)
	}
}
