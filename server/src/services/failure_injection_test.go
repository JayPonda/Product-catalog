package services_test

// Failure-injection suite: real repositories + services wired over a sqlmock
// connection, so every `if err != nil { return }` arm can be driven precisely
// (first query fails, insert fails mid-transaction, constraint violation on
// write, ...). Regex matchers key on the rendered SQL's table names.

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/src/services"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var errBoom = errors.New("boom")

var productCols = []string{"id", "name", "description", "price", "stock_quantity", "user_id", "created_at", "updated_at", "deleted_at"}
var categoryCols = []string{"id", "name", "created_at", "updated_at", "deleted_at"}
var linkCols = []string{"id", "product_id", "category_id", "created_at", "updated_at", "deleted_at"}
var userCols = []string{"id", "first_name", "last_name", "email", "password", "created_at", "updated_at", "deleted_at"}

func newMockDB(t *testing.T) (sqlmock.Sqlmock, *goqu.Database) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return mock, goqu.New("sqlite3", db)
}

func newProductServiceMock(t *testing.T) (*services.ProductService, sqlmock.Sqlmock) {
	t.Helper()
	mock, gdb := newMockDB(t)
	logger := utils.NewStructuredLogger()

	productRepo, err := repositories.InitProductRepository(gdb, logger)
	if err != nil {
		t.Fatal(err)
	}
	categoryRepo, err := repositories.InitCategoryRepository(gdb, logger)
	if err != nil {
		t.Fatal(err)
	}
	linkRepo, err := repositories.InitProductCategoryRepository(gdb, logger)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := services.InitProductService(gdb, logger, productRepo, categoryRepo, linkRepo)
	if err != nil {
		t.Fatal(err)
	}
	return svc, mock
}

func newCategoryServiceMock(t *testing.T) (*services.CategoryService, sqlmock.Sqlmock) {
	t.Helper()
	mock, gdb := newMockDB(t)
	logger := utils.NewStructuredLogger()

	categoryRepo, err := repositories.InitCategoryRepository(gdb, logger)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := services.InitCategoryService(logger, categoryRepo)
	if err != nil {
		t.Fatal(err)
	}
	return svc, mock
}

func newOrderServiceMock(t *testing.T) (*services.OrderService, sqlmock.Sqlmock) {
	t.Helper()
	mock, gdb := newMockDB(t)
	logger := utils.NewStructuredLogger()

	orderRepo, err := repositories.InitOrderRepository(gdb, logger)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := services.InitOrderService(gdb, logger, orderRepo)
	if err != nil {
		t.Fatal(err)
	}
	return svc, mock
}

func newAuthServiceMock(t *testing.T) (*services.AuthService, sqlmock.Sqlmock) {
	t.Helper()
	mock, gdb := newMockDB(t)
	logger := utils.NewStructuredLogger()

	userRepo, err := repositories.InitUserRepository(gdb, logger)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := services.InitAuthService(gdb, logger, userRepo, "injection-secret", time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return svc, mock
}

func emptyRows(cols []string) *sqlmock.Rows {
	return sqlmock.NewRows(cols)
}

func oneRow(cols []string, vals ...driver.Value) *sqlmock.Rows {
	return sqlmock.NewRows(cols).AddRow(vals...)
}

func dupViolation(constraint string) error {
	return &pq.Error{Code: "23505", Constraint: constraint}
}

func wantErr(t *testing.T, err error, name string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected injected failure, got nil", name)
	}
}

// ---- OrderService ----

func TestOrderService_ListOrders_RepoFailure(t *testing.T) {
	svc, mock := newOrderServiceMock(t)
	mock.ExpectQuery(`FROM .orders.`).WillReturnError(errBoom)

	_, err := svc.ListOrders(nil, 10, 0)
	wantErr(t, err, "ListOrders")
}

func TestOrderService_ListOrdersInRange_RepoFailure(t *testing.T) {
	svc, mock := newOrderServiceMock(t)
	mock.ExpectQuery(`FROM .orders.`).WillReturnError(errBoom)

	_, err := svc.ListOrdersInRange(nil, time.Now(), time.Now(), 10, 0)
	wantErr(t, err, "ListOrdersInRange")
}

func TestOrderService_RemoveOrder_DeleteFailure(t *testing.T) {
	svc, mock := newOrderServiceMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE .orders.`).WillReturnError(errBoom)

	if err := svc.RemoveOrder(nil, uuid.New()); err == nil {
		t.Error("expected delete failure to propagate")
	}
}

func TestOrderService_RemoveOrder_NoRows(t *testing.T) {
	svc, mock := newOrderServiceMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE .orders.`).WillReturnResult(sqlmock.NewResult(0, 0))

	if err := svc.RemoveOrder(nil, uuid.New()); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestOrderService_InTx_BeginFailure(t *testing.T) {
	svc, mock := newOrderServiceMock(t)
	mock.ExpectBegin().WillReturnError(errBoom)

	if err := svc.InTx(nil, func(tx *goqu.TxDatabase) error { return nil }); err == nil {
		t.Error("expected begin failure to propagate")
	}
}

func TestOrderService_RemoveOrders_DeleteFailure(t *testing.T) {
	svc, mock := newOrderServiceMock(t)
	mock.ExpectExec(`UPDATE .orders.`).WillReturnError(errBoom)

	if _, err := svc.RemoveOrders(nil, []uuid.UUID{uuid.New()}); err == nil {
		t.Error("expected delete failure to propagate")
	}
}

func TestOrderService_RemoveOrdersTx_DeleteFailure(t *testing.T) {
	svc, mock := newOrderServiceMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE .orders.`).WillReturnError(errBoom)

	tx, err := svc.Db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := svc.RemoveOrdersTx(nil, tx, []uuid.UUID{uuid.New()}); err == nil {
		t.Error("expected delete failure to propagate")
	}
}

// ---- ProductService ----

func seededProductRow(name string) *sqlmock.Rows {
	base := time.Now()
	return oneRow(productCols, uuid.NewString(), name, "", 100, 1, uuid.NewString(), base, base, nil)
}

func TestProductService_GetProductById_RepoFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectQuery(`FROM .products.`).WillReturnError(errBoom)

	if _, err := svc.GetProductById(nil, uuid.New()); err == nil {
		t.Error("expected lookup failure to propagate")
	}
}

func TestProductService_GetProductByName_RepoFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectQuery(`FROM .products.`).WillReturnError(errBoom)

	if _, err := svc.GetProductByName(nil, "Chair"); err == nil {
		t.Error("expected lookup failure to propagate")
	}
}

func TestProductService_ListProducts_LinksFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(seededProductRow("Chair"))
	mock.ExpectQuery(`COUNT`).WillReturnRows(oneRow([]string{"count"}, 1))
	mock.ExpectQuery(`FROM .product_categories.`).WillReturnError(errBoom)

	_, err := svc.ListProducts(nil, 20, 0, models.ProductFilter{})
	wantErr(t, err, "ListProducts links query")
}

func TestProductService_ListProducts_CategoriesFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	pid := uuid.NewString()
	base := time.Now()
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(seededProductRow("Chair"))
	mock.ExpectQuery(`COUNT`).WillReturnRows(oneRow([]string{"count"}, 1))
	mock.ExpectQuery(`FROM .product_categories.`).
		WillReturnRows(oneRow(linkCols, uuid.NewString(), pid, uuid.NewString(), base, base, nil))
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(errBoom)

	_, err := svc.ListProducts(nil, 20, 0, models.ProductFilter{})
	wantErr(t, err, "ListProducts categories query")
}

func TestProductService_CreateProduct_PreCheckFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectQuery(`FROM .products.`).WillReturnError(errBoom)

	_, err := svc.CreateProduct(nil, v1.RequestProduct{Name: "Chair"}, uuid.New())
	wantErr(t, err, "CreateProduct pre-check")
}

func TestProductService_CreateProduct_InsertFailureRollsBack(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	// Empty name lookup => not a duplicate; then the INSERT blows up.
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(emptyRows(productCols))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO .products.`).WillReturnError(errBoom)

	_, err := svc.CreateProduct(nil, v1.RequestProduct{Name: "Chair"}, uuid.New())
	wantErr(t, err, "CreateProduct insert")
}

func TestProductService_CreateProduct_ConstraintRaceMapsToSentinel(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(emptyRows(productCols))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO .products.`).WillReturnError(dupViolation("uq_products_name_active"))

	if _, err := svc.CreateProduct(nil, v1.RequestProduct{Name: "Chair"}, uuid.New()); err != services.ErrDuplicateProductName {
		t.Errorf("expected ErrDuplicateProductName, got %v", err)
	}
}

func TestProductService_UpdateProduct_PreCheckFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectQuery(`FROM .products.`).WillReturnError(errBoom)

	_, err := svc.UpdateProduct(nil, uuid.New(), v1.RequestProduct{Name: "Chair"})
	wantErr(t, err, "UpdateProduct pre-check")
}

func TestProductService_UpdateProduct_UpdateFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(emptyRows(productCols))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE .products.`).WillReturnError(errBoom)

	_, err := svc.UpdateProduct(nil, uuid.New(), v1.RequestProduct{Name: "Chair"})
	wantErr(t, err, "UpdateProduct update")
}

func linkLookupExpectations(mock sqlmock.Sqlmock, cid string) {
	base := time.Now()
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(seededProductRow("Chair"))
	mock.ExpectQuery(`FROM .categories.`).
		WillReturnRows(oneRow(categoryCols, cid, "furniture", base, base, nil))
}

func TestProductService_LinkCategory_ProductLookupFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectQuery(`FROM .products.`).WillReturnError(errBoom)

	if _, err := svc.LinkCategory(nil, uuid.New(), uuid.New()); err == nil {
		t.Error("expected product lookup failure to propagate")
	}
}

func TestProductService_LinkCategory_CategoryLookupFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	base := time.Now()
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(seededProductRow("Chair"))
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(errBoom)

	if _, err := svc.LinkCategory(nil, uuid.New(), uuid.New()); err == nil {
		_ = base
		t.Error("expected category lookup failure to propagate")
	}
}

func TestProductService_LinkCategory_LinkInsertFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	cid := uuid.NewString()
	linkLookupExpectations(mock, cid)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO .product_categories.`).WillReturnError(errBoom)

	if _, err := svc.LinkCategory(nil, uuid.New(), mustParseID(t, cid)); err == nil {
		t.Error("expected link-insert failure to propagate")
	}
}

func TestProductService_UnlinkCategory_ProductLookupFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectQuery(`FROM .products.`).WillReturnError(errBoom)

	if _, err := svc.UnlinkCategory(nil, uuid.New(), uuid.New()); err == nil {
		t.Error("expected product lookup failure to propagate")
	}
}

func TestProductService_UnlinkCategory_CategoryLookupFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(seededProductRow("Chair"))
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(errBoom)

	if _, err := svc.UnlinkCategory(nil, uuid.New(), uuid.New()); err == nil {
		t.Error("expected category lookup failure to propagate")
	}
}

func TestProductService_UnlinkCategory_UnlinkFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	cid := uuid.NewString()
	linkLookupExpectations(mock, cid)
	mock.ExpectBegin()
	mock.ExpectExec(`.product_categories.`).WillReturnError(errBoom)

	if _, err := svc.UnlinkCategory(nil, uuid.New(), mustParseID(t, cid)); err == nil {
		t.Error("expected unlink failure to propagate")
	}
}

func TestProductService_DeleteProduct_DeleteFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE .products.`).WillReturnError(errBoom)

	if err := svc.DeleteProduct(nil, uuid.New()); err == nil {
		t.Error("expected delete failure to propagate")
	}
}

func TestProductService_DeleteProduct_CascadeFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE .products.`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`.product_categories.`).WillReturnError(errBoom)

	if err := svc.DeleteProduct(nil, uuid.New()); err == nil {
		t.Error("expected cascade-delete failure to propagate")
	}
}

// ---- CategoryService ----

func TestCategoryService_ListCategories_RepoFailure(t *testing.T) {
	svc, mock := newCategoryServiceMock(t)
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(errBoom)

	_, err := svc.ListCategories(nil, 20, 0, models.CategoryFilter{})
	wantErr(t, err, "ListCategories")
}

func TestCategoryService_CreateCategory_DupCheckFailure(t *testing.T) {
	svc, mock := newCategoryServiceMock(t)
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(errBoom)

	_, err := svc.CreateCategory(nil, models.Category{Name: "Furniture"})
	wantErr(t, err, "CreateCategory dup-check")
}

func TestCategoryService_CreateCategory_InsertFailure(t *testing.T) {
	svc, mock := newCategoryServiceMock(t)
	mock.ExpectQuery(`FROM .categories.`).WillReturnRows(emptyRows(categoryCols))
	mock.ExpectExec(`INSERT INTO .categories.`).WillReturnError(errBoom)

	_, err := svc.CreateCategory(nil, models.Category{Name: "furniture"})
	wantErr(t, err, "CreateCategory insert")
}

func TestCategoryService_CreateCategory_ConstraintRaceMapsToSentinel(t *testing.T) {
	svc, mock := newCategoryServiceMock(t)
	mock.ExpectQuery(`FROM .categories.`).WillReturnRows(emptyRows(categoryCols))
	mock.ExpectExec(`INSERT INTO .categories.`).WillReturnError(dupViolation("uq_categories_name_active"))

	if _, err := svc.CreateCategory(nil, models.Category{Name: "furniture"}); err != services.ErrDuplicateCategoryName {
		t.Errorf("expected ErrDuplicateCategoryName, got %v", err)
	}
}

// ---- AuthService ----

func TestAuthService_Register_LookupFailure(t *testing.T) {
	svc, mock := newAuthServiceMock(t)
	mock.ExpectQuery(`FROM .users.`).WillReturnError(errBoom)

	_, err := svc.Register(nil, v1.RegisterRequest{Email: "a@b.c", Password: "pw"})
	wantErr(t, err, "Register lookup")
}

func TestAuthService_Register_InsertFailure(t *testing.T) {
	svc, mock := newAuthServiceMock(t)
	mock.ExpectQuery(`FROM .users.`).WillReturnRows(emptyRows(userCols))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO .users.`).WillReturnError(errBoom)

	_, err := svc.Register(nil, v1.RegisterRequest{Email: "a@b.c", Password: "pw"})
	wantErr(t, err, "Register insert")
}

func TestAuthService_Register_ConstraintRaceMapsToSentinel(t *testing.T) {
	svc, mock := newAuthServiceMock(t)
	mock.ExpectQuery(`FROM .users.`).WillReturnRows(emptyRows(userCols))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO .users.`).WillReturnError(dupViolation("uq_users_email_active"))

	if _, err := svc.Register(nil, v1.RegisterRequest{Email: "a@b.c", Password: "pw"}); err != services.ErrDuplicateEmail {
		t.Errorf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestAuthService_Login_LookupFailure(t *testing.T) {
	svc, mock := newAuthServiceMock(t)
	mock.ExpectQuery(`FROM .users.`).WillReturnError(errBoom)

	_, err := svc.Login(nil, v1.LoginRequest{Email: "a@b.c", Password: "pw"})
	wantErr(t, err, "Login lookup")
}

func TestAuthService_Login_UnknownUserIsInvalidCredentials(t *testing.T) {
	svc, mock := newAuthServiceMock(t)
	mock.ExpectQuery(`FROM .users.`).WillReturnRows(emptyRows(userCols))

	if _, err := svc.Login(nil, v1.LoginRequest{Email: "ghost@x.y", Password: "pw"}); err != services.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Logout_BeginAndDeleteFailures(t *testing.T) {
	svc, mock := newAuthServiceMock(t)
	mock.ExpectBegin().WillReturnError(errBoom)
	if err := svc.Logout(nil, uuid.New()); err == nil {
		t.Error("expected begin failure to propagate")
	}

	svc2, mock2 := newAuthServiceMock(t)
	mock2.ExpectBegin()
	mock2.ExpectExec(`.refresh_tokens.`).WillReturnError(errBoom)
	if err := svc2.Logout(nil, uuid.New()); err == nil {
		t.Error("expected revoke failure to propagate")
	}
}

func mustParseID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return id
}

// ---- Additional ProductService Failure Injection Tests ----

func TestProductService_ListMyProducts_GetMyProductsFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	uid := uuid.New()
	mock.ExpectQuery(`FROM .products.`).WillReturnError(errBoom)

	_, err := svc.ListMyProducts(nil, uid, 10, 0, models.ProductFilter{})
	wantErr(t, err, "ListMyProducts GetMyProducts")
}

func TestProductService_ListMyProducts_GetCategoriesByProductIdsFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	uid := uuid.New()
	base := time.Now()
	// GetMyProducts succeeds (returns 1 product)
	mock.ExpectQuery(`FROM .products.`).
		WillReturnRows(oneRow(productCols, uuid.NewString(), "Chair", "", 100, 1, uid.String(), base, base, nil))
	mock.ExpectQuery(`COUNT`).WillReturnRows(oneRow([]string{"count"}, 1))
	// ProductCategoryManager.GetCategoriesByProductIds fails
	mock.ExpectQuery(`FROM .product_categories.`).WillReturnError(errBoom)

	_, err := svc.ListMyProducts(nil, uid, 10, 0, models.ProductFilter{})
	wantErr(t, err, "ListMyProducts GetCategoriesByProductIds")
}

func TestProductService_ListMyProducts_GetCategoryByIdsFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	uid := uuid.New()
	pid := uuid.NewString()
	base := time.Now()
	// GetMyProducts succeeds
	mock.ExpectQuery(`FROM .products.`).
		WillReturnRows(oneRow(productCols, pid, "Chair", "", 100, 1, uid.String(), base, base, nil))
	mock.ExpectQuery(`COUNT`).WillReturnRows(oneRow([]string{"count"}, 1))
	// ProductCategoryManager.GetCategoriesByProductIds succeeds
	mock.ExpectQuery(`FROM .product_categories.`).
		WillReturnRows(oneRow(linkCols, uuid.NewString(), pid, uuid.NewString(), base, base, nil))
	// CategoryManager.GetCategoryByIds fails
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(errBoom)

	_, err := svc.ListMyProducts(nil, uid, 10, 0, models.ProductFilter{})
	wantErr(t, err, "ListMyProducts GetCategoryByIds")
}

func TestProductService_UpdateProduct_DuplicateName(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	pid := uuid.New()
	existingPid := uuid.New()
	base := time.Now()
	// Existing product with the same name exists, but with a different ID!
	mock.ExpectQuery(`FROM .products.`).
		WillReturnRows(oneRow(productCols, existingPid.String(), "Chair", "", 100, 1, uuid.NewString(), base, base, nil))

	_, err := svc.UpdateProduct(nil, pid, v1.RequestProduct{Name: "Chair"})
	if !errors.Is(err, services.ErrDuplicateProductName) {
		t.Errorf("expected ErrDuplicateProductName, got %v", err)
	}
}

func TestProductService_UpdateProduct_GetProductByNameFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	pid := uuid.New()
	mock.ExpectQuery(`FROM .products.`).WillReturnError(errBoom)

	_, err := svc.UpdateProduct(nil, pid, v1.RequestProduct{Name: "Chair"})
	wantErr(t, err, "UpdateProduct GetProductByName")
}

func TestProductService_UpdateProduct_BeginFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	pid := uuid.New()
	// Lookup name returns no rows (no duplicate name)
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(emptyRows(productCols))
	// Begin transaction fails
	mock.ExpectBegin().WillReturnError(errBoom)

	_, err := svc.UpdateProduct(nil, pid, v1.RequestProduct{Name: "Chair"})
	wantErr(t, err, "UpdateProduct Begin")
}

// ---- Additional AuthService Failure Injection Tests ----

func TestAuthService_Register_BeginFailure(t *testing.T) {
	svc, mock := newAuthServiceMock(t)
	// Email lookup returns no rows
	mock.ExpectQuery(`FROM .users.`).WillReturnRows(emptyRows(userCols))
	// Begin transaction fails
	mock.ExpectBegin().WillReturnError(errBoom)

	_, err := svc.Register(nil, v1.RegisterRequest{Email: "a@b.c", Password: "pw"})
	wantErr(t, err, "Register Begin")
}

func TestAuthService_Register_CommitFailure(t *testing.T) {
	svc, mock := newAuthServiceMock(t)
	mock.ExpectQuery(`FROM .users.`).WillReturnRows(emptyRows(userCols))
	mock.ExpectBegin()
	// Insert user succeeds
	mock.ExpectExec(`INSERT INTO .users.`).WillReturnResult(sqlmock.NewResult(1, 1))
	// GetUserById is called inside CreateUser
	base := time.Now()
	uid := uuid.NewString()
	mock.ExpectQuery(`FROM .users.`).
		WillReturnRows(oneRow(userCols, uid, "R", "T", "a@b.c", "hashedpw", base, base, nil))
	// Commit transaction fails
	mock.ExpectCommit().WillReturnError(errBoom)

	_, err := svc.Register(nil, v1.RegisterRequest{Email: "a@b.c", Password: "pw"})
	wantErr(t, err, "Register Commit")
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	svc, mock := newAuthServiceMock(t)
	hashed, _ := bcrypt.GenerateFromPassword([]byte("right-password"), bcrypt.DefaultCost)
	base := time.Now()
	mock.ExpectQuery(`FROM .users.`).
		WillReturnRows(oneRow(userCols, uuid.NewString(), "R", "T", "a@b.c", string(hashed), base, base, nil))

	// Attempt login with wrong password
	_, err := svc.Login(nil, v1.LoginRequest{Email: "a@b.c", Password: "wrong-password"})
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_BeginFailure(t *testing.T) {
	svc, mock := newAuthServiceMock(t)
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	base := time.Now()
	mock.ExpectQuery(`FROM .users.`).
		WillReturnRows(oneRow(userCols, uuid.NewString(), "R", "T", "a@b.c", string(hashed), base, base, nil))
	// Begin transaction fails
	mock.ExpectBegin().WillReturnError(errBoom)

	_, err := svc.Login(nil, v1.LoginRequest{Email: "a@b.c", Password: "pw"})
	wantErr(t, err, "Login Begin")
}

func TestAuthService_Login_CreateRefreshTokenFailure(t *testing.T) {
	svc, mock := newAuthServiceMock(t)
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	base := time.Now()
	mock.ExpectQuery(`FROM .users.`).
		WillReturnRows(oneRow(userCols, uuid.NewString(), "R", "T", "a@b.c", string(hashed), base, base, nil))
	mock.ExpectBegin()
	// Insert refresh token fails
	mock.ExpectExec(`INSERT INTO .refresh_tokens.`).WillReturnError(errBoom)

	_, err := svc.Login(nil, v1.LoginRequest{Email: "a@b.c", Password: "pw"})
	wantErr(t, err, "Login CreateRefreshToken")
}

func TestAuthService_Login_CommitFailure(t *testing.T) {
	svc, mock := newAuthServiceMock(t)
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	base := time.Now()
	mock.ExpectQuery(`FROM .users.`).
		WillReturnRows(oneRow(userCols, uuid.NewString(), "R", "T", "a@b.c", string(hashed), base, base, nil))
	mock.ExpectBegin()
	// Insert refresh token succeeds
	mock.ExpectExec(`INSERT INTO .refresh_tokens.`).WillReturnResult(sqlmock.NewResult(1, 1))
	// Commit transaction fails
	mock.ExpectCommit().WillReturnError(errBoom)

	_, err := svc.Login(nil, v1.LoginRequest{Email: "a@b.c", Password: "pw"})
	wantErr(t, err, "Login Commit")
}

// ---- Additional failure injection: tx.Commit() and InTx paths ----

func TestOrderService_InTx_FunctionFailure(t *testing.T) {
	svc, mock := newOrderServiceMock(t)
	mock.ExpectBegin()

	err := svc.InTx(nil, func(tx *goqu.TxDatabase) error {
		return errBoom
	})
	wantErr(t, err, "InTx function failure")
}

func TestOrderService_InTx_CommitFailure(t *testing.T) {
	svc, mock := newOrderServiceMock(t)
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(errBoom)

	err := svc.InTx(nil, func(tx *goqu.TxDatabase) error {
		return nil
	})
	wantErr(t, err, "InTx commit failure")
}

func TestOrderService_RemoveOrder_CommitFailure(t *testing.T) {
	svc, mock := newOrderServiceMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE .orders.`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errBoom)

	if err := svc.RemoveOrder(nil, uuid.New()); err == nil {
		t.Error("expected commit failure to propagate")
	}
}

func TestProductService_DeleteProduct_CommitFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE .products.`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`.product_categories.`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit().WillReturnError(errBoom)

	if err := svc.DeleteProduct(nil, uuid.New()); err == nil {
		t.Error("expected commit failure to propagate")
	}
}

func TestProductService_CreateProduct_CommitFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	// Name lookup returns no rows (not a duplicate)
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(emptyRows(productCols))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO .products.`).WillReturnResult(sqlmock.NewResult(1, 1))
	// Commit fails
	mock.ExpectCommit().WillReturnError(errBoom)

	_, err := svc.CreateProduct(nil, v1.RequestProduct{Name: "Chair"}, uuid.New())
	wantErr(t, err, "CreateProduct commit failure")
}

func TestProductService_UpdateProduct_CommitFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	// Name lookup returns no rows (no duplicate)
	mock.ExpectQuery(`FROM .products.`).WillReturnRows(emptyRows(productCols))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE .products.`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errBoom)

	_, err := svc.UpdateProduct(nil, uuid.New(), v1.RequestProduct{Name: "Chair"})
	wantErr(t, err, "UpdateProduct commit failure")
}

func TestProductService_LinkCategory_CommitFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	cid := uuid.NewString()
	linkLookupExpectations(mock, cid)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO .product_categories.`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(errBoom)

	if _, err := svc.LinkCategory(nil, uuid.New(), mustParseID(t, cid)); err == nil {
		t.Error("expected commit failure to propagate")
	}
}

func TestProductService_UnlinkCategory_CommitFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	cid := uuid.NewString()
	linkLookupExpectations(mock, cid)
	mock.ExpectBegin()
	mock.ExpectExec(`.product_categories.`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errBoom)

	if _, err := svc.UnlinkCategory(nil, uuid.New(), mustParseID(t, cid)); err == nil {
		t.Error("expected commit failure to propagate")
	}
}

func TestCategoryService_DeleteCategory_RepoFailure(t *testing.T) {
	svc, mock := newCategoryServiceMock(t)
	mock.ExpectExec(`UPDATE .categories.`).WillReturnError(errBoom)

	if err := svc.DeleteCategory(nil, uuid.New()); err == nil {
		t.Error("expected repo failure to propagate")
	}
}

func TestProductService_GetProductById_CategoryLookupFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	pid := uuid.NewString()
	base := time.Now()
	// Product lookup succeeds
	mock.ExpectQuery(`FROM .products.`).
		WillReturnRows(oneRow(productCols, pid, "Chair", "", 100, 1, uuid.NewString(), base, base, nil))
	// ProductCategoryManager.GetCategoriesByProduct fails
	mock.ExpectQuery(`FROM .product_categories.`).WillReturnError(errBoom)

	_, err := svc.GetProductById(nil, mustParseID(t, pid))
	wantErr(t, err, "GetProductById category lookup")
}

func TestProductService_GetProductByName_CategoryLookupFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	pid := uuid.NewString()
	base := time.Now()
	// Product lookup succeeds
	mock.ExpectQuery(`FROM .products.`).
		WillReturnRows(oneRow(productCols, pid, "Chair", "", 100, 1, uuid.NewString(), base, base, nil))
	// ProductCategoryManager.GetCategoriesByProduct fails
	mock.ExpectQuery(`FROM .product_categories.`).WillReturnError(errBoom)

	_, err := svc.GetProductByName(nil, "Chair")
	wantErr(t, err, "GetProductByName category lookup")
}

func TestProductService_GetProductById_GetCategoriesByProductIdsFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	pid := uuid.NewString()
	base := time.Now()
	// Product lookup succeeds
	mock.ExpectQuery(`FROM .products.`).
		WillReturnRows(oneRow(productCols, pid, "Chair", "", 100, 1, uuid.NewString(), base, base, nil))
	// GetCategoriesByProduct succeeds (returns a link)
	mock.ExpectQuery(`FROM .product_categories.`).
		WillReturnRows(oneRow(linkCols, uuid.NewString(), pid, uuid.NewString(), base, base, nil))
	// GetCategoryByIds fails
	mock.ExpectQuery(`FROM .categories.`).WillReturnError(errBoom)

	_, err := svc.GetProductById(nil, mustParseID(t, pid))
	wantErr(t, err, "GetProductById GetCategoryByIds failure")
}

func TestProductService_LinkCategory_BeginFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	cid := uuid.NewString()
	linkLookupExpectations(mock, cid)
	mock.ExpectBegin().WillReturnError(errBoom)

	if _, err := svc.LinkCategory(nil, uuid.New(), mustParseID(t, cid)); err == nil {
		t.Error("expected begin failure to propagate")
	}
}

func TestProductService_UnlinkCategory_BeginFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	cid := uuid.NewString()
	linkLookupExpectations(mock, cid)
	mock.ExpectBegin().WillReturnError(errBoom)

	if _, err := svc.UnlinkCategory(nil, uuid.New(), mustParseID(t, cid)); err == nil {
		t.Error("expected begin failure to propagate")
	}
}

func TestProductService_DeleteProduct_BeginFailure(t *testing.T) {
	svc, mock := newProductServiceMock(t)
	mock.ExpectBegin().WillReturnError(errBoom)

	if err := svc.DeleteProduct(nil, uuid.New()); err == nil {
		t.Error("expected begin failure to propagate")
	}
}

func TestOrderService_ListOrdersInRangeTx_RepoFailure(t *testing.T) {
	svc, mock := newOrderServiceMock(t)
	mock.ExpectBegin()
	tx, err := svc.Db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	mock.ExpectQuery(`FROM .orders.`).WillReturnError(errBoom)

	if _, err := svc.ListOrdersInRangeTx(nil, tx, time.Now(), time.Now(), 10, 0); err == nil {
		t.Error("expected repo failure to propagate")
	}
}
