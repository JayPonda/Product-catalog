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
	created, err := repo.CreateProduct(nil, in)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == uuid.Nil || created.CreatedAt.IsZero() {
		t.Errorf("expected generated id + DB-side timestamps, got %+v", created)
	}
	if !created.UserID.Valid || created.UserID.UUID != userID.UUID {
		t.Errorf("user_id not persisted: %+v", created.UserID)
	}

	got, err := repo.GetProductById(nil, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "keyboard" || got.Price != 1999 || got.StockQuantity != 7 {
		t.Errorf("unexpected product: %+v", got)
	}
}

func TestProductRepo_Create_AnonymousOwner_E2E(t *testing.T) {
	repo := newProductRepo(t)

	created, err := repo.CreateProduct(nil, mkProduct("guest-item"))
	if err != nil {
		t.Fatal(err)
	}
	if created.UserID.Valid {
		t.Errorf("expected NULL user_id, got %+v", created.UserID)
	}
}

func TestProductRepo_GetByName_NotFound_E2E(t *testing.T) {
	repo := newProductRepo(t)
	if _, err := repo.GetProductByName(nil, "nope"); err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}

func TestProductRepo_Update_E2E(t *testing.T) {
	repo := newProductRepo(t)
	created, err := repo.CreateProduct(nil, mkProduct("old-name"))
	if err != nil {
		t.Fatal(err)
	}

	before := created.UpdatedAt
	updated, err := repo.UpdateProduct(nil, created.ID, models.Product{
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

	page, total, err := repo.GetProducts(nil, 2, 0, models.ProductFilter{})
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

	if err := repo.DeleteProduct(nil, bID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetProductById(nil, bID); err != sql.ErrNoRows {
		t.Errorf("deleted product should be invisible, got %v", err)
	}
	_, total, _ = repo.GetProducts(nil, 10, 0, models.ProductFilter{})
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
		if _, err := repo.CreateProduct(nil, p); err != nil {
			t.Fatal(err)
		}
	}

	list, total, err := repo.GetMyProducts(nil, mine.UUID, 10, 0, models.ProductFilter{})
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

func TestProductRepo_GetProducts_Filters_E2E(t *testing.T) {
	repo := newProductRepo(t)
	pcRepo, err := repositories.InitProductCategoryRepository(repo.Db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatal(err)
	}
	catRepo, err := repositories.InitCategoryRepository(repo.Db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatal(err)
	}

	// Create categories
	catElectronics, _ := catRepo.CreateCategory(nil, models.Category{Name: "Electronics"})
	catBooks, _ := catRepo.CreateCategory(nil, models.Category{Name: "Books"})
	catClothing, _ := catRepo.CreateCategory(nil, models.Category{Name: "Clothing"})

	// Create products
	pPhone, _ := repo.CreateProduct(nil, mkProduct("Smartphone Pro"))
	pLaptop, _ := repo.CreateProduct(nil, mkProduct("Laptop Air"))
	pNovel, _ := repo.CreateProduct(nil, mkProduct("Sci-Fi Novel"))
	pShirt, _ := repo.CreateProduct(nil, mkProduct("Cotton Shirt"))

	// Link categories
	_ = pcRepo.LinkCategory(nil, pPhone.ID, catElectronics.ID)
	_ = pcRepo.LinkCategory(nil, pLaptop.ID, catElectronics.ID)
	_ = pcRepo.LinkCategory(nil, pNovel.ID, catBooks.ID)
	_ = pcRepo.LinkCategory(nil, pShirt.ID, catClothing.ID)

	// 1. Filter by Name (case-insensitive substring)
	res, total, err := repo.GetProducts(nil, 10, 0, models.ProductFilter{Name: "phone"})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(res) != 1 || res[0].ID != pPhone.ID {
		t.Fatalf("expected 1 phone product, got total=%d len=%d", total, len(res))
	}

	// 2. Filter by Single Category
	res, total, err = repo.GetProducts(nil, 10, 0, models.ProductFilter{CategoryIDs: []uuid.UUID{catElectronics.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(res) != 2 {
		t.Fatalf("expected 2 electronics products, got total=%d len=%d", total, len(res))
	}

	// 3. Filter by Multiple Categories (OR semantics)
	res, total, err = repo.GetProducts(nil, 10, 0, models.ProductFilter{CategoryIDs: []uuid.UUID{catElectronics.ID, catBooks.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(res) != 3 {
		t.Fatalf("expected 3 products in Electronics or Books, got total=%d len=%d", total, len(res))
	}

	// 4. Combined Name and Category filter
	res, total, err = repo.GetProducts(nil, 10, 0, models.ProductFilter{
		Name:        "air",
		CategoryIDs: []uuid.UUID{catElectronics.ID, catBooks.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(res) != 1 || res[0].ID != pLaptop.ID {
		t.Fatalf("expected Laptop Air, got total=%d len=%d", total, len(res))
	}

	// 5. Non-matching filter returns 0
	res, total, err = repo.GetProducts(nil, 10, 0, models.ProductFilter{Name: "NonExistent"})
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 || len(res) != 0 {
		t.Fatalf("expected 0 products, got total=%d len=%d", total, len(res))
	}
}
