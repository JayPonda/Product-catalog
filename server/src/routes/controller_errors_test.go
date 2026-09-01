package routes_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// These tests drive the controller error branches (400 bad request / validation,
// 404 not found) and the previously un-exercised happy path of ListMyProducts,
// using the same in-memory SQLite app as the other route tests.

func TestRoutes_ProductList_QueryErrors(t *testing.T) {
	ta := newTestApp(t)

	// Non-integer limit fails query binding -> 400.
	if res := ta.do(t, "GET", "/api/v1/products?limit=abc", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("list bad limit: status=%d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}

	// Integer but not in {20,50,100} fails validation -> 400.
	if res := ta.do(t, "GET", "/api/v1/products?limit=999", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("list invalid limit: status=%d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}

	// Negative offset fails validation -> 400.
	if res := ta.do(t, "GET", "/api/v1/products?offset=-1", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("list negative offset: status=%d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}

	// Invalid category_ids UUID fails parsing -> 400.
	if res := ta.do(t, "GET", "/api/v1/products?category_ids=not-a-uuid", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("list invalid category_ids: status=%d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}
}

func TestRoutes_ProductCreate_BodyErrors(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "create-err@example.com")

	// Missing required name -> validation 400.
	res := ta.do(t, "POST", "/api/v1/products", map[string]any{
		"description": "x", "price": 10, "stock_quantity": 1,
	}, cookie)
	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("create missing name: status=%d body=%s, want 400", res.StatusCode, b)
	}
	res.Body.Close()

	// Malformed JSON -> bind 400.
	res = ta.do(t, "POST", "/api/v1/products", json.RawMessage(`{not valid json`), cookie)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("create malformed json: status=%d, want 400", res.StatusCode)
	}
	res.Body.Close()

	// Price <= 0 fails validation -> 400.
	res = ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "Thing", "price": 0, "stock_quantity": 1,
	}, cookie)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("create zero price: status=%d, want 400", res.StatusCode)
	}
	res.Body.Close()
}

func TestRoutes_ProductUpdate_Errors(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "update-err@example.com")
	validID := uuid.New().String()

	// Invalid id -> 400 (before body bind).
	if res := ta.do(t, "PUT", "/api/v1/products/not-a-uuid", map[string]any{"name": "x", "price": 1}, cookie); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("update bad id: status=%d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}

	// Missing name -> validation 400.
	res := ta.do(t, "PUT", "/api/v1/products/"+validID, map[string]any{"price": 1}, cookie)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("update missing name: status=%d, want 400", res.StatusCode)
	}
	res.Body.Close()

	// Malformed JSON -> bind 400.
	res = ta.do(t, "PUT", "/api/v1/products/"+validID, json.RawMessage(`{bad`), cookie)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("update malformed json: status=%d, want 400", res.StatusCode)
	}
	res.Body.Close()
}

func TestRoutes_ProductLinkUnlink_BodyErrors(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "link-err@example.com")
	validID := uuid.New().String()

	// Missing category_id -> validation 400.
	for _, path := range []string{"/api/v1/products/" + validID + "/categories/link", "/api/v1/products/" + validID + "/categories/unlink"} {
		res := ta.do(t, "POST", path, map[string]any{}, cookie)
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("%s missing category_id: status=%d, want 400", path, res.StatusCode)
		}
		res.Body.Close()

		res = ta.do(t, "POST", path, json.RawMessage(`{bad`), cookie)
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("%s malformed json: status=%d, want 400", path, res.StatusCode)
		}
		res.Body.Close()
	}
}

func TestRoutes_ProductGetByName_NotFound(t *testing.T) {
	ta := newTestApp(t)

	if res := ta.do(t, "GET", "/api/v1/products/name/does-not-exist-xyz", nil, ""); res.StatusCode != http.StatusNotFound {
		t.Fatalf("get by name missing: status=%d, want 404", res.StatusCode)
	} else {
		res.Body.Close()
	}
}

func TestRoutes_ProductLinkUnlink_NotFound(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "link404@example.com")

	// Create a product so the product lookup succeeds but category is missing.
	res := ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "LinkChair", "description": "d", "price": 1, "stock_quantity": 1,
	}, cookie)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create product: status=%d", res.StatusCode)
	}
	var created struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &created)
	missingCat := uuid.New().String()

	// Link to a non-existent category -> 404 (ErrCategoryNotFound).
	res = ta.do(t, "POST", "/api/v1/products/"+created.ID+"/categories/link",
		map[string]string{"category_id": missingCat}, cookie)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("link missing category: status=%d, want 404", res.StatusCode)
	}
	res.Body.Close()

	// Unlink a non-existent link -> 404 (ErrCategoryNotFound).
	res = ta.do(t, "POST", "/api/v1/products/"+created.ID+"/categories/unlink",
		map[string]string{"category_id": missingCat}, cookie)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("unlink missing category: status=%d, want 404", res.StatusCode)
	}
	res.Body.Close()

	// Link to a non-existent product -> 404 (ErrProductNotFound).
	res = ta.do(t, "POST", "/api/v1/products/"+uuid.New().String()+"/categories/link",
		map[string]string{"category_id": missingCat}, cookie)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("link missing product: status=%d, want 404", res.StatusCode)
	}
	res.Body.Close()
}

func TestRoutes_ProductDelete_BadId(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "delete-err@example.com")

	if res := ta.do(t, "DELETE", "/api/v1/products/not-a-uuid", nil, cookie); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("delete bad id: status=%d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}
}

func TestRoutes_CategoryCreate_BodyErrors(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "catcreate-err@example.com")

	// Missing name -> validation 400.
	res := ta.do(t, "POST", "/api/v1/categories", map[string]string{}, cookie)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("category missing name: status=%d, want 400", res.StatusCode)
	}
	res.Body.Close()

	// Malformed JSON -> bind 400.
	res = ta.do(t, "POST", "/api/v1/categories", json.RawMessage(`{bad`), cookie)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("category malformed json: status=%d, want 400", res.StatusCode)
	}
	res.Body.Close()
}

func TestRoutes_CategoryListMatch_QueryErrors(t *testing.T) {
	ta := newTestApp(t)

	// ListCategories: invalid limit -> 400.
	if res := ta.do(t, "GET", "/api/v1/categories?limit=abc", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("category list bad limit: status=%d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}
	if res := ta.do(t, "GET", "/api/v1/categories?limit=999", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("category list invalid limit: status=%d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}

	// MatchCategories: missing name -> 400.
	if res := ta.do(t, "GET", "/api/v1/categories/match", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("match missing name: status=%d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}
	// MatchCategories: empty name -> 400.
	if res := ta.do(t, "GET", "/api/v1/categories/match?name=", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("match empty name: status=%d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}
	// MatchCategories: invalid limit -> 400.
	if res := ta.do(t, "GET", "/api/v1/categories/match?name=f&limit=999", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("match invalid limit: status=%d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}
}

func TestRoutes_ListMyProducts_HappyAndErrors(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "myprod@example.com")

	// Happy path (owned by this user, possibly empty) -> 200.
	var list struct {
		Products []struct {
			Name string `json:"name"`
		} `json:"products"`
		Total int64 `json:"total"`
	}
	decodeBody(t, ta.do(t, "GET", "/api/v1/my-products?limit=20&offset=0", nil, cookie), &list)
	_ = fmt.Sprintf("%v", list)

	// Bad query -> 400.
	if res := ta.do(t, "GET", "/api/v1/my-products?limit=abc", nil, cookie); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("my-products bad limit: status=%d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}

	// Create a product and confirm it shows up under my-products.
	res := ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "MyChair", "description": "d", "price": 1, "stock_quantity": 1,
	}, cookie)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create for my-products: status=%d", res.StatusCode)
	}
	res.Body.Close()

	var list2 struct {
		Total int64 `json:"total"`
	}
	decodeBody(t, ta.do(t, "GET", "/api/v1/my-products?limit=20&offset=0", nil, cookie), &list2)
	if list2.Total < 1 {
		t.Errorf("expected at least 1 product in my-products, got %d", list2.Total)
	}
}
