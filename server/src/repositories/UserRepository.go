package repositories

import (
	"database/sql"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

type UserRepository struct {
	Db     *goqu.Database
	Logger *utils.StructuredLogger
}

const USER_DB = "users"
const REFRESH_DB = "refresh_tokens"

func InitUserRepository(db *goqu.Database, logger *utils.StructuredLogger) (*UserRepository, error) {
	return &UserRepository{
		Db:     db,
		Logger: logger,
	}, nil
}

func (userRepositoryPtr *UserRepository) GetUserByEmail(ctx utils.RequestContext, email string, exec ...utils.Executor) (models.User, error) {
	var user models.User
	l := userRepositoryPtr.Logger

	db := utils.ResolveExecutor(userRepositoryPtr.Db, exec)

	found, err := db.
		From(USER_DB).
		Where(
			goqu.C("email").Eq(email),
			goqu.C("deleted_at").IsNull(),
		).
		Select(
			"id",
			"first_name",
			"last_name",
			"email",
			"password",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStruct(&user)

	if err != nil {
		l.Error(ctx, "UserRepository.go", "GetUserByEmail", "failed to query user by email", utils.LoggerMeta{"email": email}, err.Error())
		return user, err
	}

	if !found {
		l.Warn(ctx, "UserRepository.go", "GetUserByEmail", "user not found by email", utils.LoggerMeta{"email": email})
		return user, sql.ErrNoRows
	}

	return user, nil
}

func (userRepositoryPtr *UserRepository) GetUserById(ctx utils.RequestContext, id uuid.UUID, exec ...utils.Executor) (models.User, error) {
	var user models.User
	l := userRepositoryPtr.Logger

	db := utils.ResolveExecutor(userRepositoryPtr.Db, exec)

	found, err := db.
		From(USER_DB).
		Where(
			goqu.C("id").Eq(id),
			goqu.C("deleted_at").IsNull(),
		).
		Select(
			"id",
			"first_name",
			"last_name",
			"email",
			"password",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStruct(&user)

	if err != nil {
		l.Error(ctx, "UserRepository.go", "GetUserById", "failed to query user by id", utils.LoggerMeta{"id": id.String()}, err.Error())
		return user, err
	}

	if !found {
		l.Warn(ctx, "UserRepository.go", "GetUserById", "user not found by id", utils.LoggerMeta{"id": id.String()})
		return user, sql.ErrNoRows
	}

	return user, nil
}

func (userRepositoryPtr *UserRepository) CreateUser(ctx utils.RequestContext, user models.User, exec ...utils.Executor) (models.User, error) {
	l := userRepositoryPtr.Logger

	id, err := utils.GetUUID()
	if err != nil {
		l.Error(ctx, "UserRepository.go", "CreateUser", "failed to generate UUID", nil, err.Error())
		return user, err
	}

	db := utils.ResolveExecutor(userRepositoryPtr.Db, exec)

	_, err = db.Insert(USER_DB).Rows(
		goqu.Record{
			"id":         id,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"password":   user.Password,
		},
	).Executor().Exec()

	if err != nil {
		l.Error(ctx, "UserRepository.go", "CreateUser", "failed to insert user", utils.LoggerMeta{"email": user.Email}, err.Error())
		return user, err
	}

	created, err := userRepositoryPtr.GetUserById(ctx, id, exec...)
	if err != nil {
		l.Error(ctx, "UserRepository.go", "CreateUser", "failed to retrieve created user", utils.LoggerMeta{"id": id.String()}, err.Error())
		return user, err
	}

	l.Debug(ctx, "UserRepository.go", "CreateUser", "user created", utils.LoggerMeta{"id": id.String(), "email": user.Email})
	return created, nil
}

func (userRepositoryPtr *UserRepository) SoftDeleteUser(ctx utils.RequestContext, id uuid.UUID, exec ...utils.Executor) error {
	l := userRepositoryPtr.Logger

	db := utils.ResolveExecutor(userRepositoryPtr.Db, exec)

	_, err := db.Update(USER_DB).Set(
		goqu.Record{
			"deleted_at": time.Now(),
		},
	).Where(
		goqu.C("id").Eq(id),
		goqu.C("deleted_at").IsNull(),
	).Executor().Exec()

	if err != nil {
		l.Error(ctx, "UserRepository.go", "SoftDeleteUser", "failed to soft delete user", utils.LoggerMeta{"id": id.String()}, err.Error())
		return err
	}

	l.Debug(ctx, "UserRepository.go", "SoftDeleteUser", "user soft deleted", utils.LoggerMeta{"id": id.String()})
	return nil
}

func (userRepositoryPtr *UserRepository) CreateRefreshToken(ctx utils.RequestContext, token models.RefreshToken, exec ...utils.Executor) error {
	l := userRepositoryPtr.Logger

	id, err := utils.GetUUID()
	if err != nil {
		l.Error(ctx, "UserRepository.go", "CreateRefreshToken", "failed to generate UUID", nil, err.Error())
		return err
	}

	db := utils.ResolveExecutor(userRepositoryPtr.Db, exec)

	_, err = db.Insert(REFRESH_DB).Rows(
		goqu.Record{
			"id":         id,
			"user_id":    token.UserID,
			"token_hash": token.TokenHash,
			"expires_at": token.ExpiresAt,
		},
	).Executor().Exec()

	if err != nil {
		l.Error(ctx, "UserRepository.go", "CreateRefreshToken", "failed to insert refresh token", utils.LoggerMeta{"user_id": token.UserID.String()}, err.Error())
		return err
	}

	l.Debug(ctx, "UserRepository.go", "CreateRefreshToken", "refresh token created", utils.LoggerMeta{"user_id": token.UserID.String()})
	return nil
}

func (userRepositoryPtr *UserRepository) GetRefreshTokenByHash(ctx utils.RequestContext, tokenHash string, exec ...utils.Executor) (models.RefreshToken, error) {
	var token models.RefreshToken

	db := utils.ResolveExecutor(userRepositoryPtr.Db, exec)

	found, err := db.
		From(REFRESH_DB).
		Where(
			goqu.C("token_hash").Eq(tokenHash),
			goqu.C("expires_at").Gt(time.Now()),
		).
		Select(
			"id",
			"user_id",
			"token_hash",
			"expires_at",
			"created_at",
		).
		ScanStruct(&token)

	if err != nil {
		return token, err
	}

	if !found {
		return token, sql.ErrNoRows
	}

	return token, nil
}

func (userRepositoryPtr *UserRepository) DeleteRefreshTokensByUser(ctx utils.RequestContext, userID uuid.UUID, exec ...utils.Executor) error {
	db := utils.ResolveExecutor(userRepositoryPtr.Db, exec)

	_, err := db.Delete(REFRESH_DB).
		Where(goqu.C("user_id").Eq(userID)).
		Executor().Exec()

	return err
}
