package services_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/src/services"
	"github.com/JayPonda/Product-catalog/server/testdb"
	"github.com/JayPonda/Product-catalog/server/utils"
)

func newCategoryService(t *testing.T) *services.CategoryService {
	t.Helper()
	db := testdb.OpenSQLite(t)
	repo, err := repositories.InitCategoryRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatal(err)
	}
	svc, err := services.InitCategoryService(utils.NewStructuredLogger(), repo)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestCategoryService_Create_HappyPath(t *testing.T) {
	svc := newCategoryService(t)

	created, err := svc.CreateCategory(nil, models.Category{Name: "  Gadgets "})
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "gadgets" {
		t.Errorf("expected normalized name, got %q", created.Name)
	}
}

func TestCategoryService_Create_EmptyName(t *testing.T) {
	svc := newCategoryService(t)

	for _, name := range []string{"", "   ", "\t"} {
		if _, err := svc.CreateCategory(nil, models.Category{Name: name}); !errors.Is(err, services.ErrEmptyCategoryName) {
			t.Errorf("name %q: expected ErrEmptyCategoryName, got %v", name, err)
		}
	}
}

func TestCategoryService_Create_Duplicate(t *testing.T) {
	svc := newCategoryService(t)

	if _, err := svc.CreateCategory(nil, models.Category{Name: "dup"}); err != nil {
		t.Fatal(err)
	}
	// Same name, different casing/whitespace must still be caught.
	if _, err := svc.CreateCategory(nil, models.Category{Name: "  DUP "}); !errors.Is(err, services.ErrDuplicateCategoryName) {
		t.Errorf("expected ErrDuplicateCategoryName, got %v", err)
	}
}

func TestCategoryService_List_PaginationFields(t *testing.T) {
	svc := newCategoryService(t)
	for _, n := range []string{"a", "b", "c"} {
		if _, err := svc.CreateCategory(nil, models.Category{Name: n}); err != nil {
			t.Fatal(err)
		}
	}

	resp, err := svc.ListCategories(nil, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 3 || resp.Limit != 2 || resp.Offset != 1 {
		t.Errorf("unexpected meta: total=%d limit=%d offset=%d", resp.Total, resp.Limit, resp.Offset)
	}
	if len(resp.Categories) != 2 {
		t.Errorf("expected 2 categories in page, got %d", len(resp.Categories))
	}
}

func TestCategoryService_GetCategoryByNames(t *testing.T) {
	svc := newCategoryService(t)
	if _, err := svc.CreateCategory(nil, models.Category{Name: "findme"}); err != nil {
		t.Fatal(err)
	}

	found, err := svc.GetCategoryByNames(nil, []string{"FINDME"})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].Name != "findme" {
		t.Errorf("expected normalized lookup hit, got %+v", found)
	}
}

func TestCategoryService_MatchCategories(t *testing.T) {
	svc := newCategoryService(t)
	for _, n := range []string{"wood", "wool", "steel"} {
		if _, err := svc.CreateCategory(nil, models.Category{Name: n}); err != nil {
			t.Fatal(err)
		}
	}

	matches, err := svc.MatchCategories(nil, "WO", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Errorf("expected 2 prefix matches, got %d", len(matches))
	}
}

func TestCategoryService_Delete(t *testing.T) {
	svc := newCategoryService(t)
	created, err := svc.CreateCategory(nil, models.Category{Name: "doomed"})
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.DeleteCategory(nil, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetCategoryById(nil, created.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows after delete, got %v", err)
	}
}
