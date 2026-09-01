package repositories_test

// Failure-injection suite for repositories: sqlmock-backed connections so every
// `if err != nil { return }` arm and not-found branch can be driven precisely.
// Regex matchers key on the rendered SQL's table names.

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

var mockErr = errors.New("mock boom")

var (
	catCols     = []string{"id", "name", "created_at", "updated_at", "deleted_at"}
	productCols = []string{"id", "name", "description", "price", "stock_quantity", "user_id", "created_at", "updated_at", "deleted_at"}
	orderCols   = []string{"id", "customer_id", "total_bill", "created_at", "updated_at", "deleted_at"}
	linkCols    = []string{"id", "product_id", "category_id", "created_at", "updated_at", "deleted_at"}
)

func newMockRepo(t *testing.T) (sqlmock.Sqlmock, *goqu.Database) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return mock, goqu.New("sqlite3", db)
}

// ---- CategoryRepository ----

func TestCategoryRepo_GetCategoryById_QueryFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(mockErr)

	if _, err := repo.GetCategoryById(nil, uuid.New()); err == nil {
		t.Error("expected query failure to propagate")
	}
}

func TestCategoryRepo_GetCategoryByIds_QueryFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(mockErr)

	if _, err := repo.GetCategoryByIds(nil, []uuid.UUID{uuid.New()}); err == nil {
		t.Error("expected query failure to propagate")
	}
}

func TestCategoryRepo_GetCategoryByNames_QueryFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(mockErr)

	if _, err := repo.GetCategoryByNames(nil, []string{"x"}); err == nil {
		t.Error("expected query failure to propagate")
	}
}

func TestCategoryRepo_MatchCategoriesByName_QueryFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(mockErr)

	if _, err := repo.MatchCategoriesByName(nil, "x", 10); err == nil {
		t.Error("expected query failure to propagate")
	}
}

func TestCategoryRepo_GetCategories_ScanFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(mockErr)

	if _, _, err := repo.GetCategories(nil, 10, 0); err == nil {
		t.Error("expected scan failure to propagate")
	}
}

func TestCategoryRepo_GetCategories_CountFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .categories.`).WillReturnRows(sqlmock.NewRows(catCols))
	mock.ExpectQuery(`COUNT`).WillReturnError(mockErr)

	if _, _, err := repo.GetCategories(nil, 10, 0); err == nil {
		t.Error("expected count failure to propagate")
	}
}

func TestCategoryRepo_CreateCategory_InsertFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`INSERT INTO .categories.`).WillReturnError(mockErr)

	if _, err := repo.CreateCategory(nil, models.Category{Name: "furniture"}); err == nil {
		t.Error("expected insert failure to propagate")
	}
}

func TestCategoryRepo_CreateCategory_RetrieveFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`INSERT INTO .categories.`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(mockErr)

	if _, err := repo.CreateCategory(nil, models.Category{Name: "furniture"}); err == nil {
		t.Error("expected retrieve failure to propagate")
	}
}

func TestCategoryRepo_DeleteCategory_ExecFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`UPDATE .categories.`).WillReturnError(mockErr)

	if err := repo.DeleteCategory(nil, uuid.New()); err == nil {
		t.Error("expected exec failure to propagate")
	}
}

// ---- ProductCategoryRepository ----

func TestProductCategoryRepo_GetCategoriesByProduct_QueryFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .product_categories.`).WillReturnError(mockErr)

	if _, err := repo.GetCategoriesByProduct(nil, uuid.New()); err == nil {
		t.Error("expected query failure to propagate")
	}
}

func TestProductCategoryRepo_GetCategoriesByProductIds_QueryFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .product_categories.`).WillReturnError(mockErr)

	if _, err := repo.GetCategoriesByProductIds(nil, []uuid.UUID{uuid.New()}); err == nil {
		t.Error("expected query failure to propagate")
	}
}

func TestProductCategoryRepo_GetProductCategory_QueryFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .product_categories.`).WillReturnError(mockErr)

	if _, _, err := repo.GetProductCategory(nil, uuid.New(), uuid.New()); err == nil {
		t.Error("expected query failure to propagate")
	}
}

func TestProductCategoryRepo_LinkCategory_PreCheckFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .product_categories.`).WillReturnError(mockErr)

	if err := repo.LinkCategory(nil, uuid.New(), uuid.New()); err == nil {
		t.Error("expected pre-check failure to propagate")
	}
}

func TestProductCategoryRepo_LinkCategory_ReactivationExecFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductCategoryRepository(gdb, utils.NewStructuredLogger())
	pid := uuid.New()
	cid := uuid.New()
	now := time.Now()
	// GetProductCategory finds a tombstoned (deleted_at valid) row -> reactivation branch
	mock.ExpectQuery(`FROM .product_categories.`).
		WillReturnRows(sqlmock.NewRows(linkCols).AddRow(uuid.NewString(), pid.String(), cid.String(), now, now, now))
	mock.ExpectExec(`UPDATE .product_categories.`).WillReturnError(mockErr)

	if err := repo.LinkCategory(nil, pid, cid); err == nil {
		t.Error("expected reactivation exec failure to propagate")
	}
}

func TestProductCategoryRepo_LinkCategory_InsertFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductCategoryRepository(gdb, utils.NewStructuredLogger())
	pid := uuid.New()
	cid := uuid.New()
	// GetProductCategory returns no row -> insert branch
	mock.ExpectQuery(`FROM .product_categories.`).WillReturnRows(sqlmock.NewRows(linkCols))
	mock.ExpectExec(`INSERT INTO .product_categories.`).WillReturnError(mockErr)

	if err := repo.LinkCategory(nil, pid, cid); err == nil {
		t.Error("expected insert failure to propagate")
	}
}

func TestProductCategoryRepo_UnlinkCategory_ExecFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`UPDATE .product_categories.`).WillReturnError(mockErr)

	if err := repo.UnlinkCategory(nil, uuid.New(), uuid.New()); err == nil {
		t.Error("expected exec failure to propagate")
	}
}

func TestProductCategoryRepo_DeleteProductCategories_ExecFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductCategoryRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`UPDATE .product_categories.`).WillReturnError(mockErr)

	if err := repo.DeleteProductCategories(nil, uuid.New()); err == nil {
		t.Error("expected exec failure to propagate")
	}
}

// ---- UserRepository ----

func TestUserRepo_GetUserByEmail_QueryFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitUserRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .users.`).WillReturnError(mockErr)

	if _, err := repo.GetUserByEmail(nil, "a@b.c"); err == nil {
		t.Error("expected query failure to propagate")
	}
}

func TestUserRepo_GetUserById_QueryFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitUserRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .users.`).WillReturnError(mockErr)

	if _, err := repo.GetUserById(nil, uuid.New()); err == nil {
		t.Error("expected query failure to propagate")
	}
}

func TestUserRepo_CreateUser_InsertFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitUserRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`INSERT INTO .users.`).WillReturnError(mockErr)

	if _, err := repo.CreateUser(nil, models.User{Email: "a@b.c"}); err == nil {
		t.Error("expected insert failure to propagate")
	}
}

func TestUserRepo_CreateUser_RetrieveFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitUserRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`INSERT INTO .users.`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`FROM .users.`).WillReturnError(mockErr)

	if _, err := repo.CreateUser(nil, models.User{Email: "a@b.c"}); err == nil {
		t.Error("expected retrieve failure to propagate")
	}
}

func TestUserRepo_SoftDeleteUser_ExecFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitUserRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`UPDATE .users.`).WillReturnError(mockErr)

	if err := repo.SoftDeleteUser(nil, uuid.New()); err == nil {
		t.Error("expected exec failure to propagate")
	}
}

func TestUserRepo_CreateRefreshToken_InsertFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitUserRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`INSERT INTO .refresh_tokens.`).WillReturnError(mockErr)

	if err := repo.CreateRefreshToken(nil, models.RefreshToken{UserID: uuid.New()}); err == nil {
		t.Error("expected insert failure to propagate")
	}
}

func TestUserRepo_GetRefreshTokenByHash_QueryFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitUserRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .refresh_tokens.`).WillReturnError(mockErr)

	if _, err := repo.GetRefreshTokenByHash(nil, "hash"); err == nil {
		t.Error("expected query failure to propagate")
	}
}

func TestUserRepo_GetRefreshTokenByHash_NotFound(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitUserRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .refresh_tokens.`).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at"}))

	if _, err := repo.GetRefreshTokenByHash(nil, "hash"); err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

// ---- OrderRepository ----

func TestOrderRepo_ListOrders_ScanFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitOrderRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .orders.`).WillReturnError(mockErr)

	if _, _, err := repo.ListOrders(nil, 10, 0); err == nil {
		t.Error("expected scan failure to propagate")
	}
}

func TestOrderRepo_ListOrders_CountFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitOrderRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .orders.`).WillReturnRows(sqlmock.NewRows(orderCols))
	mock.ExpectQuery(`COUNT`).WillReturnError(mockErr)

	if _, _, err := repo.ListOrders(nil, 10, 0); err == nil {
		t.Error("expected count failure to propagate")
	}
}

func TestOrderRepo_ListOrdersInRange_ScanFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitOrderRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .orders.`).WillReturnError(mockErr)

	if _, _, err := repo.ListOrdersInRange(nil, time.Now(), time.Now(), 10, 0); err == nil {
		t.Error("expected scan failure to propagate")
	}
}

func TestOrderRepo_ListOrdersInRange_CountFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitOrderRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .orders.`).WillReturnRows(sqlmock.NewRows(orderCols))
	mock.ExpectQuery(`COUNT`).WillReturnError(mockErr)

	if _, _, err := repo.ListOrdersInRange(nil, time.Now(), time.Now(), 10, 0); err == nil {
		t.Error("expected count failure to propagate")
	}
}

func TestOrderRepo_DeleteOrder_ExecFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitOrderRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`UPDATE .orders.`).WillReturnError(mockErr)

	if err := repo.DeleteOrder(nil, uuid.New()); err == nil {
		t.Error("expected exec failure to propagate")
	}
}

func TestOrderRepo_DeleteOrders_ExecFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitOrderRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`UPDATE .orders.`).WillReturnError(mockErr)

	if _, err := repo.DeleteOrders(nil, []uuid.UUID{uuid.New()}); err == nil {
		t.Error("expected exec failure to propagate")
	}
}

// ---- ProductRepository ----

func TestProductRepo_GetProductById_QueryFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .products.`).WillReturnError(mockErr)

	if _, err := repo.GetProductById(nil, uuid.New()); err == nil {
		t.Error("expected query failure to propagate")
	}
}

func TestProductRepo_GetProductByName_QueryFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .products.`).WillReturnError(mockErr)

	if _, err := repo.GetProductByName(nil, "Chair"); err == nil {
		t.Error("expected query failure to propagate")
	}
}

func TestProductRepo_GetProducts_ScanFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .products.`).WillReturnError(mockErr)

	if _, _, err := repo.GetProducts(nil, 10, 0, models.ProductFilter{}); err == nil {
		t.Error("expected scan failure to propagate")
	}
}

func TestProductRepo_GetProducts_CountFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(sqlmock.NewRows(productCols))
	mock.ExpectQuery(`COUNT`).WillReturnError(mockErr)

	if _, _, err := repo.GetProducts(nil, 10, 0, models.ProductFilter{}); err == nil {
		t.Error("expected count failure to propagate")
	}
}

func TestProductRepo_CreateProduct_InsertFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`INSERT INTO .products.`).WillReturnError(mockErr)

	if _, err := repo.CreateProduct(nil, models.Product{Name: "Chair"}); err == nil {
		t.Error("expected insert failure to propagate")
	}
}

func TestProductRepo_CreateProduct_RetrieveFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`INSERT INTO .products.`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`FROM .products.`).WillReturnError(mockErr)

	if _, err := repo.CreateProduct(nil, models.Product{Name: "Chair"}); err == nil {
		t.Error("expected retrieve failure to propagate")
	}
}

func TestProductRepo_UpdateProduct_UpdateFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`UPDATE .products.`).WillReturnError(mockErr)

	if _, err := repo.UpdateProduct(nil, uuid.New(), models.Product{Name: "Chair"}); err == nil {
		t.Error("expected update failure to propagate")
	}
}

func TestProductRepo_UpdateProduct_RetrieveFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`UPDATE .products.`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`FROM .products.`).WillReturnError(mockErr)

	if _, err := repo.UpdateProduct(nil, uuid.New(), models.Product{Name: "Chair"}); err == nil {
		t.Error("expected retrieve failure to propagate")
	}
}

func TestProductRepo_DeleteProduct_ExecFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectExec(`UPDATE .products.`).WillReturnError(mockErr)

	if err := repo.DeleteProduct(nil, uuid.New()); err == nil {
		t.Error("expected exec failure to propagate")
	}
}

func TestProductRepo_GetMyProducts_ScanFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .products.`).WillReturnError(mockErr)

	if _, _, err := repo.GetMyProducts(nil, uuid.New(), 10, 0, models.ProductFilter{}); err == nil {
		t.Error("expected scan failure to propagate")
	}
}

func TestProductRepo_GetMyProducts_CountFailure(t *testing.T) {
	mock, gdb := newMockRepo(t)
	repo, _ := repositories.InitProductRepository(gdb, utils.NewStructuredLogger())
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(sqlmock.NewRows(productCols))
	mock.ExpectQuery(`COUNT`).WillReturnError(mockErr)

	if _, _, err := repo.GetMyProducts(nil, uuid.New(), 10, 0, models.ProductFilter{}); err == nil {
		t.Error("expected count failure to propagate")
	}
}
