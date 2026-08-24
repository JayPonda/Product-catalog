package services

import (
	"database/sql"
	"errors"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Db              *goqu.Database
	Logger          *utils.StructuredLogger
	UserManager     *repositories.UserRepository
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func InitAuthService(
	db *goqu.Database,
	logger *utils.StructuredLogger,
	userManager *repositories.UserRepository,
	jwtSecret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) (*AuthService, error) {
	return &AuthService{
		Db:              db,
		Logger:          logger,
		UserManager:     userManager,
		JWTSecret:       jwtSecret,
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
	}, nil
}

// Register creates a new user with a bcrypt-hashed password.
func (authServicePtr *AuthService) Register(req v1.RegisterRequest) (v1.AuthResponse, error) {
	var response v1.AuthResponse

	if _, err := authServicePtr.UserManager.GetUserByEmail(req.Email); err == nil {
		return response, ErrDuplicateEmail
	} else if !errors.Is(err, sql.ErrNoRows) {
		return response, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return response, err
	}

	tx, err := authServicePtr.Db.Begin()
	if err != nil {
		return response, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	user, err := authServicePtr.UserManager.CreateUser(models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  string(hashed),
	}, tx)
	if err != nil {
		if IsDuplicateEmail(err) {
			return response, ErrDuplicateEmail
		}
		return response, err
	}

	if err := tx.Commit(); err != nil {
		return response, err
	}

	response.User = v1.ToUserResponse(user)
	return response, nil
}

// LoginResult holds the tokens issued on successful authentication.
type LoginResult struct {
	User         models.User
	AccessToken  string
	RefreshToken string
}

// Login verifies credentials and issues an access + refresh token pair.
func (authServicePtr *AuthService) Login(req v1.LoginRequest) (LoginResult, error) {
	var result LoginResult

	user, err := authServicePtr.UserManager.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, ErrInvalidCredentials
		}
		return result, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return result, ErrInvalidCredentials
	}

	accessToken, err := utils.GenerateAccessToken(user.ID.String(), authServicePtr.JWTSecret, authServicePtr.AccessTokenTTL)
	if err != nil {
		return result, err
	}

	rawRefresh, hashRefresh, err := utils.GenerateRefreshToken()
	if err != nil {
		return result, err
	}

	tx, err := authServicePtr.Db.Begin()
	if err != nil {
		return result, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := authServicePtr.UserManager.CreateRefreshToken(models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashRefresh,
		ExpiresAt: time.Now().Add(authServicePtr.RefreshTokenTTL),
	}, tx); err != nil {
		return result, err
	}

	if err := tx.Commit(); err != nil {
		return result, err
	}

	result.User = user
	result.AccessToken = accessToken
	result.RefreshToken = rawRefresh
	return result, nil
}

// Logout revokes all refresh tokens for the user (effective logout).
func (authServicePtr *AuthService) Logout(userID uuid.UUID) error {
	tx, err := authServicePtr.Db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := authServicePtr.UserManager.DeleteRefreshTokensByUser(userID, tx); err != nil {
		return err
	}

	return tx.Commit()
}
