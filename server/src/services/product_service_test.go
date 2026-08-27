package services_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/src/services"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/testdb"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/google/uuid"
)

// newProductService wires the full production service stack on a real
// in-memory SQLite database.
func newProductService(t *testing.T) (*services.ProductService, *repositories.CategoryRepository) {
	t.Helper()
	db := testdb.OpenSQLite(t)

	productRepo, err := repositories.InitProductRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatal(err)
	}
	categoryRepo, err := repositories.InitCategoryRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatal(err)
	}
	pcRepo, err := repositories.InitProductCategoryRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatal(err)
	}
	svc, err := services.InitProductService(db, utils.NewStructuredLogger(), productRepo, categoryRepo, pcRepo)
	if err != nil {
		t.Fatal(err)
	}
	return svc, categoryRepo
}

// seedProductWithCategory creates a product owned by userID plus one
// category, returning both ids.
func seedProductWithCategory(
	t *testing.T,
	svc *services.ProductService,
	categoryRepo *repositories.CategoryRepository,
	userID uuid.UUID,
) (uuid.UUID, uuid.UUID) {
	t.Helper()

	created, err := svc.CreateProduct(nil, v1.RequestProduct{
		Name:          "Widget",
		Description:   "a widget",
		Price:         1999,
		StockQuantity: 5,
	}, userID)
	if err != nil {
		t.Fatal(err)
	}

	cat, err := categoryRepo.CreateCategory(nil, models.Category{Name: "tools"})
	if err != nil {
		t.Fatal(err)
	}
	return created.Product.ID, cat.ID
}

func TestProductService_Create_HappyPath_E2E(t *testing.T) {
	svc, categoryRepo := newProductService(t)
	userID := uuid.New()

	productID, _ := seedProductWithCategory(t, svc, categoryRepo, userID)

	got, err := svc.GetProductById(nil, productID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Product.Name != "Widget" || got.Product.Price != 1999 {
		t.Errorf("unexpected product: %+v", got.Product)
	}
	if !got.Product.UserID.Valid || got.Product.UserID.UUID != userID {
		t.Errorf("expected owner set, got %+v", got.Product.UserID)
	}
}

func TestProductService_Create_DuplicateName_E2E(t *testing.T) {
	svc, categoryRepo := newProductService(t)
	userID := uuid.New()
	seedProductWithCategory(t, svc, categoryRepo, userID)

	dup := v1.RequestProduct{Name: "Widget", Price: 999, StockQuantity: 1}
	if _, err := svc.CreateProduct(nil, dup, uuid.New()); !errors.Is(err, services.ErrDuplicateProductName) {
		t.Errorf("expected ErrDuplicateProductName, got %v", err)
	}
}

func TestProductService_GetProductById_WithCategories_E2E(t *testing.T) {
	svc, categoryRepo := newProductService(t)
	userID := uuid.New()
	productID, catID := seedProductWithCategory(t, svc, categoryRepo, userID)

	if _, err := svc.LinkCategory(nil, productID, catID); err != nil {
		t.Fatal(err)
	}

	got, err := svc.GetProductById(nil, productID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Categories) != 1 || got.Categories[0].Name != "tools" {
		t.Errorf("expected [tools], got %+v", got.Categories)
	}
}

func TestProductService_ListProducts_HydratesCategories_E2E(t *testing.T) {
	svc, categoryRepo := newProductService(t)
	userID := uuid.New()
	productID, catID := seedProductWithCategory(t, svc, categoryRepo, userID)

	if _, err := svc.LinkCategory(nil, productID, catID); err != nil {
		t.Fatal(err)
	}
	// Second product owns no categories.
	second, err := svc.CreateProduct(nil, v1.RequestProduct{Name: "Gizmo", Price: 1, StockQuantity: 2}, userID)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := svc.ListProducts(nil, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 2 || resp.Limit != 10 || resp.Offset != 0 {
		t.Fatalf("unexpected meta: %+v", resp)
	}

	catsByProduct := map[uuid.UUID]int{}
	for _, item := range resp.Products {
		catsByProduct[item.Product.ID] = len(item.Categories)
	}
	if catsByProduct[productID] != 1 {
		t.Errorf("widget should carry 1 category, got %d", catsByProduct[productID])
	}
	if catsByProduct[second.Product.ID] != 0 {
		t.Errorf("gizmo should carry 0 categories, got %d", catsByProduct[second.Product.ID])
	}
}

func TestProductService_UpdateProduct_Rename_E2E(t *testing.T) {
	svc, categoryRepo := newProductService(t)
	productID, _ := seedProductWithCategory(t, svc, categoryRepo, uuid.New())

	updated, err := svc.UpdateProduct(nil, productID, v1.RequestProduct{
		Name: "Widget Pro", Description: "bigger", Price: 2999, StockQuantity: 7,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Product.Name != "Widget Pro" || updated.Product.StockQuantity != 7 {
		t.Errorf("update not applied: %+v", updated.Product)
	}
}

func TestProductService_UpdateProduct_KeepOwnName_E2E(t *testing.T) {
	svc, categoryRepo := newProductService(t)
	productID, _ := seedProductWithCategory(t, svc, categoryRepo, uuid.New())

	// Keeping the same name must NOT count as a duplicate.
	updated, err := svc.UpdateProduct(nil, productID, v1.RequestProduct{
		Name: "Widget", Price: 21.00, StockQuantity: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Product.Price != 21.00 {
		t.Errorf("price not updated: %v", updated.Product.Price)
	}
}

func TestProductService_UpdateProduct_NameCollision_E2E(t *testing.T) {
	svc, categoryRepo := newProductService(t)
	userID := uuid.New()
	seedProductWithCategory(t, svc, categoryRepo, userID)

	other, err := svc.CreateProduct(nil, v1.RequestProduct{Name: "Taken", Price: 2, StockQuantity: 1}, userID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.UpdateProduct(nil, other.Product.ID, v1.RequestProduct{
		Name: "Widget", Price: 3, StockQuantity: 1,
	})
	if !errors.Is(err, services.ErrDuplicateProductName) {
		t.Errorf("expected ErrDuplicateProductName, got %v", err)
	}
}

func TestProductService_LinkUnlink_Errors_E2E(t *testing.T) {
	svc, categoryRepo := newProductService(t)
	productID, catID := seedProductWithCategory(t, svc, categoryRepo, uuid.New())

	if _, err := svc.LinkCategory(nil, uuid.New(), catID); !errors.Is(err, services.ErrProductNotFound) {
		t.Errorf("link unknown product: expected ErrProductNotFound, got %v", err)
	}
	if _, err := svc.LinkCategory(nil, productID, uuid.New()); !errors.Is(err, services.ErrCategoryNotFound) {
		t.Errorf("link unknown category: expected ErrCategoryNotFound, got %v", err)
	}
	if _, err := svc.UnlinkCategory(nil, uuid.New(), catID); !errors.Is(err, services.ErrProductNotFound) {
		t.Errorf("unlink unknown product: expected ErrProductNotFound, got %v", err)
	}
	if _, err := svc.UnlinkCategory(nil, productID, uuid.New()); !errors.Is(err, services.ErrCategoryNotFound) {
		t.Errorf("unlink unknown category: expected ErrCategoryNotFound, got %v", err)
	}
}

func TestProductService_Unlink_RemovesFromResponse_E2E(t *testing.T) {
	svc, categoryRepo := newProductService(t)
	userID := uuid.New()
	productID, catID := seedProductWithCategory(t, svc, categoryRepo, userID)

	if _, err := svc.LinkCategory(nil, productID, catID); err != nil {
		t.Fatal(err)
	}
	got, err := svc.UnlinkCategory(nil, productID, catID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Categories) != 0 {
		t.Errorf("expected no categories after unlink, got %+v", got.Categories)
	}
}

func TestProductService_Delete_CascadesLinks_E2E(t *testing.T) {
	svc, categoryRepo := newProductService(t)
	userID := uuid.New()
	productID, catID := seedProductWithCategory(t, svc, categoryRepo, userID)

	if _, err := svc.LinkCategory(nil, productID, catID); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteProduct(nil, productID); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.GetProductById(nil, productID); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows after delete, got %v", err)
	}

	// Link row must be tombstoned, not left active.
	link, found, err := svc.ProductCategoryManager.GetProductCategory(nil, productID, catID)
	if err != nil || !found {
		t.Fatalf("link row missing entirely: found=%v err=%v", found, err)
	}
	if !link.DeletedAt.Valid {
		t.Error("link row still active after product delete")
	}
}

func TestProductService_ListMyProducts_Isolated_E2E(t *testing.T) {
	svc, categoryRepo := newProductService(t)
	userID := uuid.New()
	seedProductWithCategory(t, svc, categoryRepo, userID)

	otherUser := uuid.New()
	if _, err := svc.CreateProduct(nil, v1.RequestProduct{Name: "Theirs", Price: 1, StockQuantity: 1}, otherUser); err != nil {
		t.Fatal(err)
	}

	mine, err := svc.ListMyProducts(nil, userID, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if mine.Total != 1 || len(mine.Products) != 1 {
		t.Fatalf("expected exactly 1 owned product, got total=%d", mine.Total)
	}
	if mine.Products[0].Product.Name != "Widget" {
		t.Errorf("wrong product returned: %+v", mine.Products[0].Product.Name)
	}
}
