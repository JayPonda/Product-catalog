package repositories_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/testdb"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

// ---- shared helpers ----

func newOrderRepoE2E(t *testing.T) *repositories.OrderRepository {
	t.Helper()
	db := testdb.OpenSQLite(t)
	repo, err := repositories.InitOrderRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatalf("init order repo: %v", err)
	}
	return repo
}

func newOrderRepoMock(t *testing.T, db *goqu.Database) *repositories.OrderRepository {
	t.Helper()
	return &repositories.OrderRepository{Db: db, Logger: utils.NewStructuredLogger()}
}

func mustUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", s, err)
	}
	return id
}

// ---- E2E: real SQLite, zero mocking ----

func TestOrderRepo_ListOrdersInRange_E2E(t *testing.T) {
	repo := newOrderRepoE2E(t)
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	custA := uuid.NewString()
	custB := uuid.NewString()

	testdb.SeedOrder(t, repo.Db, uuid.NewString(), custA, 100.0, base.Add(-time.Minute))
	testdb.SeedOrder(t, repo.Db, uuid.NewString(), custA, 100.0, base.Add(time.Minute))
	testdb.SeedOrder(t, repo.Db, uuid.NewString(), custB, 200.0, base.Add(10*time.Minute))

	orders, total, err := repo.ListOrdersInRange(nil, base.Add(-time.Hour), base.Add(time.Hour), 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(orders) != 3 {
		t.Fatalf("expected 3 orders (total %d, got %d)", total, len(orders))
	}
	if !orders[0].CreatedAt.Before(orders[1].CreatedAt) || !orders[1].CreatedAt.Before(orders[2].CreatedAt) {
		t.Error("expected orders ordered by created_at ASC")
	}

	// Range that excludes custA's two orders (kept >5m away from them).
	orders2, total2, err := repo.ListOrdersInRange(nil, base.Add(5*time.Minute), base.Add(time.Hour), 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total2 != 1 || len(orders2) != 1 || orders2[0].CustomerID.String() != custB {
		t.Errorf("expected only custB in narrowed range, got total=%d len=%d", total2, len(orders2))
	}
}

func TestOrderRepo_DeleteOrders_E2E(t *testing.T) {
	repo := newOrderRepoE2E(t)
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	id1 := uuid.NewString()
	id2 := uuid.NewString()
	testdb.SeedOrder(t, repo.Db, id1, uuid.NewString(), 1.0, base)
	testdb.SeedOrder(t, repo.Db, id2, uuid.NewString(), 1.0, base.Add(time.Minute))

	affected, err := repo.DeleteOrders(nil, []uuid.UUID{mustUUID(t, id1), mustUUID(t, id2)})
	if err != nil {
		t.Fatal(err)
	}
	if affected != 2 {
		t.Errorf("affected = %d, want 2", affected)
	}

	orders, total, err := repo.ListOrdersInRange(nil, base.Add(-time.Hour), base.Add(time.Hour), 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 || len(orders) != 0 {
		t.Errorf("expected 0 orders after soft-delete, got total=%d", total)
	}
}

func TestOrderRepo_DeleteOrder_E2E(t *testing.T) {
	repo := newOrderRepoE2E(t)
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	id := uuid.NewString()
	testdb.SeedOrder(t, repo.Db, id, uuid.NewString(), 1.0, base)

	if err := repo.DeleteOrder(nil, mustUUID(t, id)); err != nil {
		t.Fatalf("first delete: %v", err)
	}
	if err := repo.DeleteOrder(nil, mustUUID(t, id)); err != sql.ErrNoRows {
		t.Errorf("second delete should be ErrNoRows, got %v", err)
	}
}

func TestOrderRepo_ListOrders_Pagination_E2E(t *testing.T) {
	repo := newOrderRepoE2E(t)
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		testdb.SeedOrder(t, repo.Db, uuid.NewString(), uuid.NewString(), float64(i),
			base.Add(time.Duration(i)*time.Minute))
	}

	page1, total, err := repo.ListOrders(nil, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 5 || len(page1) != 2 {
		t.Errorf("page1: total=%d len=%d, want 5/2", total, len(page1))
	}
	page2, _, err := repo.ListOrders(nil, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2) != 2 {
		t.Errorf("page2 len = %d, want 2", len(page2))
	}
}

// ---- Unit: in-memory sqlmock (no real DB) ----

func TestOrderRepo_ListOrdersInRange_Sqlmock(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := newOrderRepoMock(t, goqu.New("sqlite3", db))

	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "customer_id", "total_bill", "created_at", "updated_at", "deleted_at"}).
		AddRow(uuid.NewString(), uuid.NewString(), 100.0, base, base, nil).
		AddRow(uuid.NewString(), uuid.NewString(), 100.0, base.Add(time.Minute), base, nil)
	mock.ExpectQuery(`ORDER BY`).WillReturnRows(rows)
	mock.ExpectQuery(`COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	orders, total, err := repo.ListOrdersInRange(nil, base.Add(-time.Hour), base.Add(time.Hour), 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(orders) != 2 {
		t.Errorf("got total=%d len=%d, want 2/2", total, len(orders))
	}
	if orders[0].TotalBill != 100.0 || orders[1].TotalBill != 100.0 {
		t.Errorf("unexpected total_bill mapping: %v", orders)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestOrderRepo_DeleteOrders_Sqlmock(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := newOrderRepoMock(t, goqu.New("sqlite3", db))

	id1 := uuid.NewString()
	id2 := uuid.NewString()
	mock.ExpectExec(`deleted_at`).
		WillReturnResult(sqlmock.NewResult(0, 2))

	affected, err := repo.DeleteOrders(nil, []uuid.UUID{mustUUID(t, id1), mustUUID(t, id2)})
	if err != nil {
		t.Fatal(err)
	}
	if affected != 2 {
		t.Errorf("affected = %d, want 2", affected)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// compile-time guard: ensure seeded model shape is consistent.
var _ = models.Order{}
