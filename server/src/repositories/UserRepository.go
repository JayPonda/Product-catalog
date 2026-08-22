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

func (userRepositoryPtr *UserRepository) GetUserByEmail(email string, exec ...utils.Executor) (models.User, error) {
	var user models.User

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
		return user, err
	}

	if !found {
		return user, sql.ErrNoRows
	}

	return user, nil
}

func (userRepositoryPtr *UserRepository) GetUserById(id uuid.UUID, exec ...utils.Executor) (models.User, error) {
	var user models.User

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
		return user, err
	}

	if !found {
		return user, sql.ErrNoRows
	}

	return user, nil
}

func (userRepositoryPtr *UserRepository) CreateUser(user models.User, exec ...utils.Executor) (models.User, error) {
	id, err := utils.GetUUID()
	if err != nil {
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
		return user, err
	}

	return userRepositoryPtr.GetUserById(id, exec...)
}

func (userRepositoryPtr *UserRepository) SoftDeleteUser(id uuid.UUID, exec ...utils.Executor) error {
	db := utils.ResolveExecutor(userRepositoryPtr.Db, exec)

	_, err := db.Update(USER_DB).Set(
		goqu.Record{
			"deleted_at": time.Now(),
		},
	).Where(
		goqu.C("id").Eq(id),
		goqu.C("deleted_at").IsNull(),
	).Executor().Exec()

	return err
}

func (userRepositoryPtr *UserRepository) CreateRefreshToken(token models.RefreshToken, exec ...utils.Executor) error {
	id, err := utils.GetUUID()
	if err != nil {
		return err
	}

	db := utils.ResolveExecutor(userRepositoryPtr.Db, exec)

	_, err = db.Insert(REFRESH_DB).Rows(
		goqu.Record{
			"id":          id,
			"user_id":     token.UserID,
			"token_hash":  token.TokenHash,
			"expires_at":  token.ExpiresAt,
		},
	).Executor().Exec()

	return err
}

func (userRepositoryPtr *UserRepository) GetRefreshTokenByHash(tokenHash string, exec ...utils.Executor) (models.RefreshToken, error) {
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

func (userRepositoryPtr *UserRepository) DeleteRefreshTokensByUser(userID uuid.UUID, exec ...utils.Executor) error {
	db := utils.ResolveExecutor(userRepositoryPtr.Db, exec)

	_, err := db.Delete(REFRESH_DB).
		Where(goqu.C("user_id").Eq(userID)).
		Executor().Exec()

	return err
}
