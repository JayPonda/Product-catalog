package repositories_test

import (
	"database/sql"
	"testing"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/testdb"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/google/uuid"
)

func newCategoryRepo(t *testing.T) *repositories.CategoryRepository {
	t.Helper()
	db := testdb.OpenSQLite(t)
	repo, err := repositories.InitCategoryRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatalf("init category repo: %v", err)
	}
	return repo
}

func seedCategories(t *testing.T, repo *repositories.CategoryRepository, names ...string) []models.Category {
	t.Helper()
	out := make([]models.Category, 0, len(names))
	for _, n := range names {
		c, err := repo.CreateCategory(nil, models.Category{Name: n})
		if err != nil {
			t.Fatalf("create category %q: %v", n, err)
		}
		out = append(out, c)
	}
	return out
}

func TestCategoryRepo_Create_NormalizesName_E2E(t *testing.T) {
	repo := newCategoryRepo(t)

	created, err := repo.CreateCategory(nil, models.Category{Name: "  Electronics  "})
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "electronics" {
		t.Errorf("expected normalized name %q, got %q", "electronics", created.Name)
	}
	if created.ID == uuid.Nil || created.CreatedAt.IsZero() {
		t.Errorf("expected id + DB-side timestamps, got %+v", created)
	}
}

func TestCategoryRepo_GetById_And_NotFound_E2E(t *testing.T) {
	repo := newCategoryRepo(t)
	seeded := seedCategories(t, repo, "toys")

	got, err := repo.GetCategoryById(nil, seeded[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "toys" {
		t.Errorf("expected toys, got %q", got.Name)
	}
	if _, err := repo.GetCategoryById(nil, uuid.New()); err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows for unknown id, got %v", err)
	}
}

func TestCategoryRepo_GetCategories_Pagination_E2E(t *testing.T) {
	repo := newCategoryRepo(t)
	seedCategories(t, repo, "alpha", "beta", "gamma", "delta")

	page, total, err := repo.GetCategories(nil, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 4 || len(page) != 2 {
		t.Fatalf("page1 total=%d len=%d, want 4/2", total, len(page))
	}
	// ordered by name ASC
	if page[0].Name != "alpha" || page[1].Name != "beta" {
		t.Errorf("expected alpha,beta got %q,%q", page[0].Name, page[1].Name)
	}

	page2, _, _ := repo.GetCategories(nil, 2, 2)
	if len(page2) != 2 || page2[0].Name != "delta" || page2[1].Name != "gamma" {
		t.Errorf("unexpected page2: %v", page2)
	}
}

func TestCategoryRepo_GetByNames_Normalized_E2E(t *testing.T) {
	repo := newCategoryRepo(t)
	seedCategories(t, repo, "sports", "outdoor")

	found, err := repo.GetCategoryByNames(nil, []string{"  SPORTS ", "", "outdoor"})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(found))
	}
	names := map[string]bool{found[0].Name: true, found[1].Name: true}
	if !names["sports"] || !names["outdoor"] {
		t.Errorf("expected sports+outdoor, got %v", names)
	}
}

func TestCategoryRepo_GetByIds_Empty_E2E(t *testing.T) {
	repo := newCategoryRepo(t)
	seedCategories(t, repo, "books", "games")

	found, err := repo.GetCategoryByIds(nil, []uuid.UUID{})
	if err != nil {
		t.Fatalf("unexpected error on empty ids: %v", err)
	}
	if len(found) != 0 {
		t.Fatalf("expected 0 categories, got %d", len(found))
	}
}

func TestCategoryRepo_GetByNames_Empty_E2E(t *testing.T) {
	repo := newCategoryRepo(t)
	seedCategories(t, repo, "books", "games")

	found, err := repo.GetCategoryByNames(nil, []string{})
	if err != nil {
		t.Fatalf("unexpected error on empty names: %v", err)
	}
	if len(found) != 0 {
		t.Fatalf("expected 0 categories, got %d", len(found))
	}

	// All blank / whitespace names that normalize to empty
	foundBlank, err := repo.GetCategoryByNames(nil, []string{"", "   "})
	if err != nil {
		t.Fatalf("unexpected error on blank names: %v", err)
	}
	if len(foundBlank) != 0 {
		t.Fatalf("expected 0 categories for blanks, got %d", len(foundBlank))
	}
}

func TestCategoryRepo_MatchByNamePrefix_E2E(t *testing.T) {
	repo := newCategoryRepo(t)
	seedCategories(t, repo, "shoes", "shirts", "shorts", "hats")

	matches, err := repo.MatchCategoriesByName(nil, "SH", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(matches))
	}
	for _, m := range matches {
		if m.Name != "shoes" && m.Name != "shirts" && m.Name != "shorts" {
			t.Errorf("unexpected match %q", m.Name)
		}
	}

	limited, _ := repo.MatchCategoriesByName(nil, "sh", 2)
	if len(limited) != 2 {
		t.Errorf("limit not applied, got %d", len(limited))
	}

	none, _ := repo.MatchCategoriesByName(nil, "zzz", 10)
	if len(none) != 0 {
		t.Errorf("expected zero matches, got %d", len(none))
	}
}

func TestCategoryRepo_Delete_HidesCategory_E2E(t *testing.T) {
	repo := newCategoryRepo(t)
	seeded := seedCategories(t, repo, "temp")

	if err := repo.DeleteCategory(nil, seeded[0].ID); err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}
	if _, err := repo.GetCategoryById(nil, seeded[0].ID); err != sql.ErrNoRows {
		t.Errorf("deleted category should be invisible, got %v", err)
	}
	_, total, _ := repo.GetCategories(nil, 10, 0)
	if total != 0 {
		t.Errorf("deleted category still counted, total=%d", total)
	}
}
