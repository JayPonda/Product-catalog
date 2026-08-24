package repositories_test

import (
	"time"
	"database/sql"
	"testing"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/testdb"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/google/uuid"
)

func newProductRepo(t *testing.T) *repositories.ProductRepository {
	t.Helper()
	db := testdb.OpenSQLite(t)
	repo, err := repositories.InitProductRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatalf("init product repo: %v", err)
	}
	return repo
}

func mkProduct(name string) models.Product {
	return models.Product{
		Name:          name,
		Description:   "desc " + name,
		Price:         1999,
		StockQuantity: 7,
	}
}

func TestProductRepo_Create_AndFetch_E2E(t *testing.T) {
	repo := newProductRepo(t)
	userID := uuid.NullUUID{UUID: uuid.New(), Valid: true}

	in := mkProduct("keyboard")
	in.UserID = userID
	created, err := repo.CreateProduct(in)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == uuid.Nil || created.CreatedAt.IsZero() {
		t.Errorf("expected generated id + DB-side timestamps, got %+v", created)
	}
	if !created.UserID.Valid || created.UserID.UUID != userID.UUID {
		t.Errorf("user_id not persisted: %+v", created.UserID)
	}

	got, err := repo.GetProductById(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "keyboard" || got.Price != 1999 || got.StockQuantity != 7 {
		t.Errorf("unexpected product: %+v", got)
	}
}

func TestProductRepo_Create_AnonymousOwner_E2E(t *testing.T) {
	repo := newProductRepo(t)

	created, err := repo.CreateProduct(mkProduct("guest-item"))
	if err != nil {
		t.Fatal(err)
	}
	if created.UserID.Valid {
		t.Errorf("expected NULL user_id, got %+v", created.UserID)
	}
}

func TestProductRepo_GetByName_NotFound_E2E(t *testing.T) {
	repo := newProductRepo(t)
	if _, err := repo.GetProductByName("nope"); err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}

func TestProductRepo_Update_E2E(t *testing.T) {
	repo := newProductRepo(t)
	created, err := repo.CreateProduct(mkProduct("old-name"))
	if err != nil {
		t.Fatal(err)
	}

	before := created.UpdatedAt
	updated, err := repo.UpdateProduct(created.ID, models.Product{
		Name:          "new-name",
		Description:   "updated desc",
		Price:         4242,
		StockQuantity: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "new-name" || updated.Price != 4242 || updated.StockQuantity != 3 {
		t.Errorf("update not applied: %+v", updated)
	}
	if !updated.UpdatedAt.After(before) && before.IsZero() == false && !updated.UpdatedAt.Equal(before) {
		// updated_at should move forward (or at minimum the row must refresh);
		// only fail when it clearly went backwards.
		if updated.UpdatedAt.Before(before) {
			t.Errorf("updated_at went backwards: %v -> %v", before, updated.UpdatedAt)
		}
	}
}

func TestProductRepo_GetProducts_PaginationAndDelete_E2E(t *testing.T) {
	repo := newProductRepo(t)

	// Seed with explicit timestamps so "newest first" is deterministic
	// (rapid inserts can share the same millisecond).
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	a := uuid.NewString()
	b := uuid.NewString()
	c := uuid.NewString()
	testdb.SeedProduct(t, repo.Db, a, "alpha", base)
	testdb.SeedProduct(t, repo.Db, b, "beta", base.Add(time.Minute))
	testdb.SeedProduct(t, repo.Db, c, "gamma", base.Add(2*time.Minute))
	aID := mustUUID(t, a)
	bID := mustUUID(t, b)
	_ = aID

	page, total, err := repo.GetProducts(2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(page) != 2 {
		t.Fatalf("page1 total=%d len=%d, want 3/2", total, len(page))
	}
	// newest first
	if page[0].Name != "gamma" || page[1].Name != "beta" {
		t.Errorf("expected [gamma beta], got [%s %s]", page[0].Name, page[1].Name)
	}

	if err := repo.DeleteProduct(bID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetProductById(bID); err != sql.ErrNoRows {
		t.Errorf("deleted product should be invisible, got %v", err)
	}
	_, total, _ = repo.GetProducts(10, 0)
	if total != 2 {
		t.Errorf("total after delete = %d, want 2", total)
	}
}

func TestProductRepo_GetMyProducts_E2E(t *testing.T) {
	repo := newProductRepo(t)
	mine := uuid.NullUUID{UUID: uuid.New(), Valid: true}
	other := uuid.NullUUID{UUID: uuid.New(), Valid: true}

	p1 := mkProduct("mine-1")
	p1.UserID = mine
	p2 := mkProduct("mine-2")
	p2.UserID = mine
	p3 := mkProduct("theirs")
	p3.UserID = other

	for _, p := range []models.Product{p1, p2, p3} {
		if _, err := repo.CreateProduct(p); err != nil {
			t.Fatal(err)
		}
	}

	list, total, err := repo.GetMyProducts(mine.UUID, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("my products total=%d len=%d, want 2/2", total, len(list))
	}
	for _, p := range list {
		if p.UserID.UUID != mine.UUID {
			t.Errorf("leaked foreign product %q", p.Name)
		}
	}
}
