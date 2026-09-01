package repositories_test

import (
	"testing"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/testdb"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/google/uuid"
)

func newProductCategoryRepo(t *testing.T) (*repositories.ProductCategoryRepository, *repositories.CategoryRepository) {
	t.Helper()
	db := testdb.OpenSQLite(t)
	pcRepo, err := repositories.InitProductCategoryRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatalf("init product-category repo: %v", err)
	}
	catRepo, err := repositories.InitCategoryRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatalf("init category repo: %v", err)
	}
	return pcRepo, catRepo
}

// linkFixture creates one category and links it to a fresh product id.
func linkFixture(t *testing.T, pcRepo *repositories.ProductCategoryRepository, catRepo *repositories.CategoryRepository) (uuid.UUID, uuid.UUID) {
	t.Helper()
	cat, err := catRepo.CreateCategory(nil, models.Category{Name: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	productID := uuid.New()
	if err := pcRepo.LinkCategory(nil, productID, cat.ID); err != nil {
		t.Fatalf("LinkCategory: %v", err)
	}
	return productID, cat.ID
}

func TestProductCategoryRepo_Link_AndFetch_E2E(t *testing.T) {
	pcRepo, catRepo := newProductCategoryRepo(t)
	productID, catID := linkFixture(t, pcRepo, catRepo)

	links, err := pcRepo.GetCategoriesByProduct(nil, productID)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].CategoryID != catID || links[0].ProductID != productID {
		t.Fatalf("unexpected links: %+v", links)
	}
	byIds, err := pcRepo.GetCategoriesByProductIds(nil, []uuid.UUID{productID, uuid.New()})
	if err != nil {
		t.Fatal(err)
	}
	if len(byIds) != 1 {
		t.Errorf("expected 1 link via ids query, got %d", len(byIds))
	}
	found, ok, err := pcRepo.GetProductCategory(nil, productID, catID)
	if err != nil || !ok {
		t.Fatalf("expected link found (ok=%v err=%v)", ok, err)
	}
	if found.DeletedAt.Valid {
		t.Error("freshly linked row should be active")
	}
}

func TestProductCategoryRepo_GetCategoriesByProductIds_Empty_E2E(t *testing.T) {
	pcRepo, _ := newProductCategoryRepo(t)
	byIds, err := pcRepo.GetCategoriesByProductIds(nil, []uuid.UUID{})
	if err != nil {
		t.Fatalf("unexpected error on empty productIDs: %v", err)
	}
	if len(byIds) != 0 {
		t.Errorf("expected 0 links via empty ids query, got %d", len(byIds))
	}
}

func TestProductCategoryRepo_Unlink_E2E(t *testing.T) {
	pcRepo, catRepo := newProductCategoryRepo(t)
	productID, catID := linkFixture(t, pcRepo, catRepo)

	if err := pcRepo.UnlinkCategory(nil, productID, catID); err != nil {
		t.Fatal(err)
	}
	links, _ := pcRepo.GetCategoriesByProduct(nil, productID)
	if len(links) != 0 {
		t.Errorf("unlinked category still visible: %+v", links)
	}

	// GetProductCategory sees the soft-deleted row (found=true, DeletedAt valid).
	row, ok, err := pcRepo.GetProductCategory(nil, productID, catID)
	if err != nil || !ok {
		t.Fatalf("expected tombstoned row visible to GetProductCategory (ok=%v err=%v)", ok, err)
	}
	if !row.DeletedAt.Valid {
		t.Error("expected deleted_at set on unlinked row")
	}
}

func TestProductCategoryRepo_Relink_Reactivates_E2E(t *testing.T) {
	pcRepo, catRepo := newProductCategoryRepo(t)
	productID, catID := linkFixture(t, pcRepo, catRepo)

	if err := pcRepo.UnlinkCategory(nil, productID, catID); err != nil {
		t.Fatal(err)
	}
	// Re-linking must reactivate the tombstoned relationship, not duplicate it.
	if err := pcRepo.LinkCategory(nil, productID, catID); err != nil {
		t.Fatal(err)
	}

	links, _ := pcRepo.GetCategoriesByProduct(nil, productID)
	if len(links) != 1 {
		t.Fatalf("expected exactly 1 reactivated link, got %d", len(links))
	}
	if links[0].DeletedAt.Valid {
		t.Error("reactivated link should have NULL deleted_at")
	}
	_ = catRepo
}

func TestProductCategoryRepo_DeleteAllForProduct_E2E(t *testing.T) {
	pcRepo, catRepo := newProductCategoryRepo(t)

	c1, err := catRepo.CreateCategory(nil, models.Category{Name: "one"})
	if err != nil {
		t.Fatal(err)
	}
	c2, err := catRepo.CreateCategory(nil, models.Category{Name: "two"})
	if err != nil {
		t.Fatal(err)
	}
	productID := uuid.New()
	other := uuid.New()
	for _, cid := range []uuid.UUID{c1.ID, c2.ID} {
		if err := pcRepo.LinkCategory(nil, productID, cid); err != nil {
			t.Fatal(err)
		}
		if err := pcRepo.LinkCategory(nil, other, cid); err != nil {
			t.Fatal(err)
		}
	}

	if err := pcRepo.DeleteProductCategories(nil, productID); err != nil {
		t.Fatal(err)
	}

	mine, _ := pcRepo.GetCategoriesByProduct(nil, productID)
	if len(mine) != 0 {
		t.Errorf("expected all links for product removed, got %d", len(mine))
	}
	othersLinks, _ := pcRepo.GetCategoriesByProduct(nil, other)
	if len(othersLinks) != 2 {
		t.Errorf("other product's links must be untouched, got %d", len(othersLinks))
	}
}
