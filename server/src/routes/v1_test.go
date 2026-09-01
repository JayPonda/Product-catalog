package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	controllersv1 "github.com/JayPonda/Product-catalog/server/src/controllers/v1"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/src/routes"
	"github.com/JayPonda/Product-catalog/server/src/services"
	"github.com/JayPonda/Product-catalog/server/testdb"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3"
)

const testJWTSecret = "routes-test-secret"

type testApp struct {
	app         *fiber.App
	authService *services.AuthService
	db          *goqu.Database
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	db := testdb.OpenSQLite(t)
	logger := utils.NewStructuredLogger()

	productRepo, err := repositories.InitProductRepository(db, logger)
	if err != nil {
		t.Fatal(err)
	}
	categoryRepo, err := repositories.InitCategoryRepository(db, logger)
	if err != nil {
		t.Fatal(err)
	}
	pcRepo, err := repositories.InitProductCategoryRepository(db, logger)
	if err != nil {
		t.Fatal(err)
	}
	userRepo, err := repositories.InitUserRepository(db, logger)
	if err != nil {
		t.Fatal(err)
	}
	orderRepo, err := repositories.InitOrderRepository(db, logger)
	if err != nil {
		t.Fatal(err)
	}

	productSvc, err := services.InitProductService(db, logger, productRepo, categoryRepo, pcRepo)
	if err != nil {
		t.Fatal(err)
	}
	categorySvc, err := services.InitCategoryService(logger, categoryRepo)
	if err != nil {
		t.Fatal(err)
	}
	authSvc, err := services.InitAuthService(db, logger, userRepo, testJWTSecret, 15*60*1e9, 24*3600*1e9)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = services.InitOrderService(db, logger, orderRepo)

	app := fiber.New()
	routes.RegisterV1Routes(
		app,
		controllersv1.NewProductController(productSvc, logger),
		controllersv1.NewCategoryController(categorySvc, logger),
		controllersv1.NewAuthController(authSvc, false, logger),
		testJWTSecret,
		logger,
	)

	return &testApp{app: app, authService: authSvc, db: db}
}

// do executes a request against the app and returns the response.
func (ta *testApp) do(t *testing.T, method, path string, body any, cookie string) *http.Response {
	t.Helper()

	var rdr io.Reader
	if body != nil {
		// json.RawMessage is passed through verbatim so tests can send
		// deliberately malformed bodies at the bind layer.
		if raw, ok := body.(json.RawMessage); ok {
			rdr = bytes.NewReader(raw)
		} else {
			buf, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			rdr = bytes.NewReader(buf)
		}
	}

	req := httptest.NewRequest(method, path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	res, err := ta.app.Test(req, fiber.TestConfig{Timeout: 10_000_000_000})
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return res
}

func decodeBody(t *testing.T, res *http.Response, into any) {
	t.Helper()
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(into); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

// loginAs registers (if needed), logs in, and returns a Cookie header value.
func (ta *testApp) loginAs(t *testing.T, email string) string {
	t.Helper()

	res := ta.do(t, "POST", "/api/v1/auth/register", map[string]string{
		"first_name": "R", "last_name": "T", "email": email, "password": "password123",
	}, "")
	res.Body.Close()

	res = ta.do(t, "POST", "/api/v1/auth/login", map[string]string{
		"email": email, "password": "password123",
	}, "")
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("login %s: status=%d body=%s", email, res.StatusCode, b)
	}

	parts := []string{}
	for _, c := range res.Cookies() {
		parts = append(parts, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	res.Body.Close()
	return parts[0]
}

func TestRoutes_AuthFlow_E2E(t *testing.T) {
	ta := newTestApp(t)

	res := ta.do(t, "POST", "/api/v1/auth/register", map[string]string{
		"first_name": "Ada", "last_name": "L", "email": "ada@example.com", "password": "password123",
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("register: status=%d", res.StatusCode)
	}
	var reg struct {
		User struct{ Email string } `json:"user"`
	}
	decodeBody(t, res, &reg)
	if reg.User.Email != "ada@example.com" {
		t.Errorf("wrong user in response: %+v", reg.User)
	}

	// Duplicate registration → 409.
	if res = ta.do(t, "POST", "/api/v1/auth/register", map[string]string{
		"first_name": "Ada", "last_name": "L", "email": "ada@example.com", "password": "password123",
	}, ""); res.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate register: status=%d, want 409", res.StatusCode)
	}
	res.Body.Close()

	// Bad payload → 400.
	if res = ta.do(t, "POST", "/api/v1/auth/register", map[string]string{
		"first_name": "Ada", "email": "not-an-email", "password": "short",
	}, ""); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid register: status=%d, want 400", res.StatusCode)
	}
	res.Body.Close()

	// Login sets both cookies.
	res = ta.do(t, "POST", "/api/v1/auth/login", map[string]string{
		"email": "ada@example.com", "password": "password123",
	}, "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("login: status=%d", res.StatusCode)
	}
	names := map[string]bool{}
	for _, c := range res.Cookies() {
		names[c.Name] = true
	}
	if !names["access_token"] || !names["refresh_token"] {
		t.Errorf("expected auth cookies, got %v", names)
	}

	// Wrong password → 401.
	if res = ta.do(t, "POST", "/api/v1/auth/login", map[string]string{
		"email": "ada@example.com", "password": "wrongpass99",
	}, ""); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("bad login: status=%d, want 401", res.StatusCode)
	}
	res.Body.Close()
}

func TestRoutes_Me_Guarded_E2E(t *testing.T) {
	ta := newTestApp(t)

	// No cookie → 401.
	if res := ta.do(t, "GET", "/api/v1/auth/me", nil, ""); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated me: status=%d, want 401", res.StatusCode)
	} else {
		res.Body.Close()
	}

	email := "me@example.com"
	loginRes := ta.do(t, "POST", "/api/v1/auth/register", map[string]string{
		"first_name": "M", "last_name": "E", "email": email, "password": "password123",
	}, "")
	loginRes.Body.Close()

	cookie := ta.loginAs(t, email)

	var me struct {
		User struct{ Email string } `json:"user"`
	}
	decodeBody(t, ta.do(t, "GET", "/api/v1/auth/me", nil, cookie), &me)
	if me.User.Email != email {
		t.Errorf("me returned wrong user: %+v", me.User)
	}

	// Garbage token → 401.
	if res := ta.do(t, "GET", "/api/v1/auth/me", nil, "access_token=forged.jwt.token"); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("forged token: status=%d, want 401", res.StatusCode)
	} else {
		res.Body.Close()
	}
}

func TestRoutes_ProductCRUD_GuardsAndCodes_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "owner@example.com")

	// Unauthenticated create → 401.
	if res := ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "Nope", "price": 100, "stock_quantity": 1,
	}, ""); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauth create: status=%d, want 401", res.StatusCode)
	} else {
		res.Body.Close()
	}

	payload := map[string]any{"name": "Chair", "description": "a chair", "price": 4999, "stock_quantity": 10}
	res := ta.do(t, "POST", "/api/v1/products", payload, cookie)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create: status=%d", res.StatusCode)
	}
	var created struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Price int64  `json:"price"`
	}
	decodeBody(t, res, &created)
	if created.Name != "Chair" || created.Price != 4999 {
		t.Errorf("unexpected product: %+v", created)
	}

	// Duplicate name → 409.
	if res = ta.do(t, "POST", "/api/v1/products", payload, cookie); res.StatusCode != http.StatusConflict {
		t.Fatalf("dup create: status=%d, want 409", res.StatusCode)
	}
	res.Body.Close()

	// Public list works without auth.
	var list struct {
		Products []struct {
			Name string `json:"name"`
		} `json:"products"`
		Total int64 `json:"total"`
	}
	decodeBody(t, ta.do(t, "GET", "/api/v1/products?limit=20&offset=0", nil, ""), &list)
	if list.Total < 1 || len(list.Products) < 1 {
		t.Errorf("list empty: total=%d", list.Total)
	}

	// Get by id / name / unknown.
	var one struct {
		ID string `json:"id"`
	}
	decodeBody(t, ta.do(t, "GET", "/api/v1/products/"+created.ID, nil, ""), &one)
	if one.ID != created.ID {
		t.Errorf("get by id mismatch: %+v vs %s", one, created.ID)
	}
	if res = ta.do(t, "GET", "/api/v1/products/name/Chair", nil, ""); res.StatusCode != http.StatusOK {
		t.Fatalf("get by name: status=%d", res.StatusCode)
	}
	res.Body.Close()
	if res = ta.do(t, "GET", "/api/v1/products/not-a-uuid", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad id: status=%d, want 400", res.StatusCode)
	}
	res.Body.Close()
	if res = ta.do(t, "GET", fmt.Sprintf("/api/v1/products/%s", uuidZero()), nil, ""); res.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown id: status=%d, want 404", res.StatusCode)
	}
	res.Body.Close()

	// Update via API.
	res = ta.do(t, "PUT", "/api/v1/products/"+created.ID, map[string]any{
		"name": "Chair Deluxe", "description": "comfy", "price": 5999, "stock_quantity": 7,
	}, cookie)
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("update: status=%d body=%s", res.StatusCode, b)
	}
	var updated struct {
		Name  string `json:"name"`
		Price int64  `json:"price"`
	}
	decodeBody(t, res, &updated)
	if updated.Name != "Chair Deluxe" || updated.Price != 5999 {
		t.Errorf("update not applied: %+v", updated)
	}

	// Delete → then gone.
	if res = ta.do(t, "DELETE", "/api/v1/products/"+created.ID, nil, cookie); res.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: status=%d, want 204", res.StatusCode)
	}
	res.Body.Close()
	if res = ta.do(t, "DELETE", "/api/v1/products/"+created.ID, nil, cookie); res.StatusCode == http.StatusInternalServerError || res.StatusCode == http.StatusNotImplemented {
		t.Fatalf("double delete: unexpected status=%d", res.StatusCode)
	} else {
		res.Body.Close()
	}
	if res = ta.do(t, "GET", "/api/v1/products/name/Chair%20Deluxe", nil, ""); res.StatusCode != http.StatusNotFound {
		t.Fatalf("deleted product still fetchable: status=%d", res.StatusCode)
	}
	res.Body.Close()
}

func TestRoutes_CategoryCreateMatch_LinkUnlink_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "cat@example.com")

	// Create category (normalized server-side).
	res := ta.do(t, "POST", "/api/v1/categories", map[string]string{"name": "  Furniture "}, cookie)
	if res.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("category create: status=%d body=%s", res.StatusCode, b)
	}
	var cat struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	decodeBody(t, res, &cat)
	if cat.Name != "furniture" {
		t.Errorf("name not normalized: %q", cat.Name)
	}

	// Duplicate → 409.
	if res = ta.do(t, "POST", "/api/v1/categories", map[string]string{"name": "Furniture"}, cookie); res.StatusCode != http.StatusConflict {
		t.Fatalf("dup category: status=%d, want 409", res.StatusCode)
	}
	res.Body.Close()

	// Prefix match is public.
	var match struct {
		Categories []struct {
			Name string `json:"name"`
		} `json:"categories"`
	}
	decodeBody(t, ta.do(t, "GET", "/api/v1/categories/match?name=FURN&limit=20", nil, ""), &match)
	if len(match.Categories) != 1 || match.Categories[0].Name != "furniture" {
		t.Errorf("match results: %+v", match.Categories)
	}

	// Create a product and link the category through the API.
	res = ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "Table", "description": "a table", "price": 19900, "stock_quantity": 3,
	}, cookie)
	var created struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &created)

	res = ta.do(t, "POST", "/api/v1/products/"+created.ID+"/categories/link",
		map[string]string{"category_id": cat.ID}, cookie)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("link: status=%d", res.StatusCode)
	}
	var linked struct {
		Categories []struct {
			Name string `json:"name"`
		} `json:"categories"`
	}
	decodeBody(t, res, &linked)
	if len(linked.Categories) != 1 || linked.Categories[0].Name != "furniture" {
		t.Errorf("link not reflected: %+v", linked.Categories)
	}

	res = ta.do(t, "POST", "/api/v1/products/"+created.ID+"/categories/unlink",
		map[string]string{"category_id": cat.ID}, cookie)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("unlink: status=%d", res.StatusCode)
	}
	decodeBody(t, res, &linked)
	if len(linked.Categories) != 0 {
		t.Errorf("unlink not reflected: %+v", linked.Categories)
	}

	// Link with unknown category → 404. (The all-zero UUID would fail
	// validator "required" earlier and return 400 instead.)
	if res = ta.do(t, "POST", "/api/v1/products/"+created.ID+"/categories/link",
		map[string]string{"category_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7"}, cookie); res.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown link: status=%d, want 404", res.StatusCode)
	}
	res.Body.Close()
}

func TestRoutes_MyProducts_ScopedToOwner_E2E(t *testing.T) {
	ta := newTestApp(t)

	alice := ta.loginAs(t, "alice@example.com")
	bob := ta.loginAs(t, "bob@example.com")

	res := ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "Alice Item", "description": "hers", "price": 100, "stock_quantity": 1,
	}, alice)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("alice create: status=%d", res.StatusCode)
	}
	res.Body.Close()

	var bobList struct {
		Total int64 `json:"total"`
	}
	decodeBody(t, ta.do(t, "GET", "/api/v1/my-products", nil, bob), &bobList)
	if bobList.Total != 0 {
		t.Errorf("bob sees %d products, want 0", bobList.Total)
	}

	var aliceList struct {
		Products []struct {
			Name string `json:"name"`
		} `json:"products"`
		Total int64 `json:"total"`
	}
	decodeBody(t, ta.do(t, "GET", "/api/v1/my-products", nil, alice), &aliceList)
	if aliceList.Total != 1 || aliceList.Products[0].Name != "Alice Item" {
		t.Errorf("alice list wrong: total=%d %+v", aliceList.Total, aliceList.Products)
	}

	// Unauthenticated my-products → 401.
	if res := ta.do(t, "GET", "/api/v1/my-products", nil, ""); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauth my-products: status=%d, want 401", res.StatusCode)
	} else {
		res.Body.Close()
	}

	// Bad query params → 400
	if res := ta.do(t, "GET", "/api/v1/my-products?limit=7", nil, alice); res.StatusCode != http.StatusBadRequest {
		t.Errorf("my-products bad limit: %d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}
}

func TestRoutes_Logout_RevokesAndClearsCookies_E2E(t *testing.T) {
	ta := newTestApp(t)
	email := "out@example.com"
	cookie := ta.loginAs(t, email)

	// Unauthenticated logout → 401.
	if res := ta.do(t, "POST", "/api/v1/auth/logout", nil, ""); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauth logout: status=%d, want 401", res.StatusCode)
	} else {
		res.Body.Close()
	}

	res := ta.do(t, "POST", "/api/v1/auth/logout", nil, cookie)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("logout: status=%d, want 204", res.StatusCode)
	}
	cleared := map[string]bool{}
	for _, c := range res.Cookies() {
		if c.Value == "" && c.Expires.Before(time.Now()) {
			cleared[c.Name] = true
		}
	}
	if !cleared["access_token"] || !cleared["refresh_token"] {
		t.Errorf("expected both auth cookies cleared, got %v", cleared)
	}
}

func TestRoutes_CategoryList_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "listcat@example.com")

	for _, name := range []string{"Toys", "Furniture"} {
		res := ta.do(t, "POST", "/api/v1/categories", map[string]string{"name": name}, cookie)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("create %q: status=%d, want 201", name, res.StatusCode)
		}
	}

	var list struct {
		Categories []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"categories"`
		Total int64 `json:"total"`
		Limit int   `json:"limit"`
	}
	decodeBody(t, ta.do(t, "GET", "/api/v1/categories?limit=20&offset=0", nil, ""), &list)
	if list.Total != 2 || len(list.Categories) != 2 {
		t.Fatalf("total=%d len=%d, want 2/2", list.Total, len(list.Categories))
	}
	if list.Categories[0].Name != "furniture" || list.Categories[1].Name != "toys" {
		t.Errorf("expected alphabetical order furniture,toys; got %s,%s",
			list.Categories[0].Name, list.Categories[1].Name)
	}

	// Omitted params hit the default-limit branch.
	defaultList := list
	defaultList.Categories = nil
	decodeBody(t, ta.do(t, "GET", "/api/v1/categories", nil, ""), &defaultList)
	if defaultList.Limit != 20 || defaultList.Total != 2 {
		t.Errorf("defaults: limit=%d total=%d, want 20/2", defaultList.Limit, defaultList.Total)
	}

	// limit must be one of 20|50|100.
	if res := ta.do(t, "GET", "/api/v1/categories?limit=7", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Errorf("limit=7: status=%d, want 400", res.StatusCode)
	}
}

func uuidZero() string {
	return "00000000-0000-0000-0000-000000000000"
}

// ---- bind / validation 400s and service-failure mappings ----

// dropTable makes every query against `table` fail, forcing the service layer
// into its error arms so the controller's status-code mapping can be asserted.
func dropTable(t *testing.T, ta *testApp, table string) {
	t.Helper()
	if _, err := ta.db.Exec("DROP TABLE IF EXISTS " + table); err != nil {
		t.Fatalf("drop %s: %v", table, err)
	}
}

func TestRoutes_BindAndValidation_400s_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "bind@example.com")

	// Malformed JSON bodies hit the ctx.Bind().JSON arm.
	if res := ta.do(t, "POST", "/api/v1/auth/register", json.RawMessage("{oops"), ""); res.StatusCode != http.StatusBadRequest {
		t.Errorf("register malformed: %d, want 400", res.StatusCode)
	}
	if res := ta.do(t, "POST", "/api/v1/auth/login", json.RawMessage("{oops"), ""); res.StatusCode != http.StatusBadRequest {
		t.Errorf("login malformed: %d, want 400", res.StatusCode)
	}
	if res := ta.do(t, "POST", "/api/v1/categories", json.RawMessage("{oops"), cookie); res.StatusCode != http.StatusBadRequest {
		t.Errorf("categories malformed: %d, want 400", res.StatusCode)
	}
	if res := ta.do(t, "POST", "/api/v1/products", json.RawMessage("{oops"), cookie); res.StatusCode != http.StatusBadRequest {
		t.Errorf("products malformed: %d, want 400", res.StatusCode)
	}

	// Struct validation arms.
	if res := ta.do(t, "POST", "/api/v1/auth/login", map[string]string{"email": "a@b.c"}, ""); res.StatusCode != http.StatusBadRequest {
		t.Errorf("login missing password: %d, want 400", res.StatusCode)
	}
	if res := ta.do(t, "GET", "/api/v1/categories/match?limit=20", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Errorf("match without name: %d, want 400", res.StatusCode)
	}
	if res := ta.do(t, "GET", "/api/v1/categories/match?name=FURN&limit=7", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Errorf("match bad limit: %d, want 400", res.StatusCode)
	}
	if res := ta.do(t, "GET", "/api/v1/products?limit=7", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Errorf("products bad limit: %d, want 400", res.StatusCode)
	}
	if res := ta.do(t, "GET", "/api/v1/categories?offset=-3", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Errorf("categories negative offset: %d, want 400", res.StatusCode)
	}
}

func TestRoutes_ServiceFailures_MapToErrorStatuses_E2E(t *testing.T) {
	t.Run("products table gone", func(t *testing.T) {
		ta := newTestApp(t)
		cookie := ta.loginAs(t, "drop-p@example.com")
		dropTable(t, ta, "products")

		if res := ta.do(t, "GET", "/api/v1/products", nil, ""); res.StatusCode != http.StatusInternalServerError {
			t.Errorf("list products: %d, want 500", res.StatusCode)
		}
		// GetProductByName maps every service error to 404 by contract.
		if res := ta.do(t, "GET", "/api/v1/products/name/Chair", nil, ""); res.StatusCode != http.StatusNotFound {
			t.Errorf("product by name: %d, want 404", res.StatusCode)
		}
		res := ta.do(t, "POST", "/api/v1/products", map[string]any{
			"name": "Chair", "description": "d", "price": 100, "stock_quantity": 1,
		}, cookie)
		if res.StatusCode != http.StatusInternalServerError {
			t.Errorf("create product: %d, want 500", res.StatusCode)
		}
	})

	t.Run("categories table gone", func(t *testing.T) {
		ta := newTestApp(t)
		cookie := ta.loginAs(t, "drop-c@example.com")
		dropTable(t, ta, "categories")

		if res := ta.do(t, "GET", "/api/v1/categories", nil, ""); res.StatusCode != http.StatusInternalServerError {
			t.Errorf("list categories: %d, want 500", res.StatusCode)
		}
		if res := ta.do(t, "GET", "/api/v1/categories/match?name=FURN&limit=5", nil, ""); res.StatusCode != http.StatusInternalServerError {
			t.Errorf("match categories: %d, want 500", res.StatusCode)
		}
		if res := ta.do(t, "POST", "/api/v1/categories", map[string]string{"name": "furniture"}, cookie); res.StatusCode != http.StatusInternalServerError {
			t.Errorf("create category: %d, want 500", res.StatusCode)
		}
	})

	t.Run("users table gone", func(t *testing.T) {
		ta := newTestApp(t)
		email := "drop-u@example.com"
		cookie := ta.loginAs(t, email)
		dropTable(t, ta, "users")

		if res := ta.do(t, "POST", "/api/v1/auth/register", map[string]string{
			"first_name": "A", "last_name": "B", "email": "new@x.y", "password": "password123",
		}, ""); res.StatusCode != http.StatusInternalServerError {
			t.Errorf("register: %d, want 500", res.StatusCode)
		}
		if res := ta.do(t, "POST", "/api/v1/auth/login", map[string]string{"email": email, "password": "pw"}, ""); res.StatusCode != http.StatusInternalServerError {
			t.Errorf("login: %d, want 500", res.StatusCode)
		}
		// /auth/me maps a vanished user to 401 (not 500).
		if res := ta.do(t, "GET", "/api/v1/auth/me", nil, cookie); res.StatusCode != http.StatusUnauthorized {
			t.Errorf("me after users dropped: %d, want 401", res.StatusCode)
		}
	})

	t.Run("refresh_tokens table gone", func(t *testing.T) {
		ta := newTestApp(t)
		email := "drop-r@example.com"
		cookie := ta.loginAs(t, email)
		dropTable(t, ta, "refresh_tokens")

		// Login writes the refresh token inside its transaction.
		if res := ta.do(t, "POST", "/api/v1/auth/login", map[string]string{"email": email, "password": "password123"}, ""); res.StatusCode != http.StatusInternalServerError {
			t.Errorf("login with refresh gone: %d, want 500", res.StatusCode)
		}
		if res := ta.do(t, "POST", "/api/v1/auth/logout", nil, cookie); res.StatusCode != http.StatusInternalServerError {
			t.Errorf("logout with refresh gone: %d, want 500", res.StatusCode)
		}
	})
}

func TestRoutes_ProductUpdate_ErrorPaths_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "update-err@example.com")

	// Create a product first
	res := ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "ToUpdate", "description": "d", "price": 100, "stock_quantity": 1,
	}, cookie)
	var created struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &created)

	// Update with invalid UUID → 400
	if res = ta.do(t, "PUT", "/api/v1/products/not-a-uuid", map[string]any{
		"name": "X", "description": "d", "price": 100, "stock_quantity": 1,
	}, cookie); res.StatusCode != http.StatusBadRequest {
		t.Errorf("update bad uuid: %d, want 400", res.StatusCode)
	}
	res.Body.Close()

	// Update with malformed JSON → 400
	if res = ta.do(t, "PUT", "/api/v1/products/"+created.ID, json.RawMessage("{oops"), cookie); res.StatusCode != http.StatusBadRequest {
		t.Errorf("update malformed: %d, want 400", res.StatusCode)
	}
	res.Body.Close()

	// Update with missing name → 400
	if res = ta.do(t, "PUT", "/api/v1/products/"+created.ID, map[string]any{
		"name": "", "description": "d", "price": 100, "stock_quantity": 1,
	}, cookie); res.StatusCode != http.StatusBadRequest {
		t.Errorf("update empty name: %d, want 400", res.StatusCode)
	}
	res.Body.Close()

	// Update non-existent product → service error (no ErrProductNotFound mapping in controller)
	zeroID := "00000000-0000-0000-0000-000000000000"
	if res = ta.do(t, "PUT", "/api/v1/products/"+zeroID, map[string]any{
		"name": "X", "description": "d", "price": 100, "stock_quantity": 1,
	}, cookie); res.StatusCode != http.StatusInternalServerError {
		t.Errorf("update non-existent: %d, want 500", res.StatusCode)
	}
	res.Body.Close()

	// Update with duplicate name → 409
	res2 := ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "ExistingProd", "description": "d", "price": 100, "stock_quantity": 1,
	}, cookie)
	var created2 struct {
		ID string `json:"id"`
	}
	decodeBody(t, res2, &created2)
	if res = ta.do(t, "PUT", "/api/v1/products/"+created.ID, map[string]any{
		"name": "ExistingProd", "description": "d", "price": 100, "stock_quantity": 1,
	}, cookie); res.StatusCode != http.StatusConflict {
		t.Errorf("update dup name: %d, want 409", res.StatusCode)
	}
	res.Body.Close()
}

func TestRoutes_ProductDelete_ErrorPaths_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "del-err@example.com")

	// Delete with invalid UUID → 400
	if res := ta.do(t, "DELETE", "/api/v1/products/not-a-uuid", nil, cookie); res.StatusCode != http.StatusBadRequest {
		t.Errorf("delete bad uuid: %d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}
}

func TestRoutes_ProductLinkUnlink_ErrorPaths_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "link-err@example.com")

	// Create a product
	res := ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "LinkTest", "description": "d", "price": 100, "stock_quantity": 1,
	}, cookie)
	var created struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &created)

	// Link with invalid product UUID → 400
	if res = ta.do(t, "POST", "/api/v1/products/not-a-uuid/categories/link",
		map[string]string{"category_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7"}, cookie); res.StatusCode != http.StatusBadRequest {
		t.Errorf("link bad product uuid: %d, want 400", res.StatusCode)
	}
	res.Body.Close()

	// Link with malformed JSON → 400
	if res = ta.do(t, "POST", "/api/v1/products/"+created.ID+"/categories/link",
		json.RawMessage("{oops"), cookie); res.StatusCode != http.StatusBadRequest {
		t.Errorf("link malformed: %d, want 400", res.StatusCode)
	}
	res.Body.Close()

	// Link with non-existent category → 404
	fakeCatID := "7c9e6679-7425-40de-944b-e07fc1f90ae7"
	if res = ta.do(t, "POST", "/api/v1/products/"+created.ID+"/categories/link",
		map[string]string{"category_id": fakeCatID}, cookie); res.StatusCode != http.StatusNotFound {
		t.Errorf("link non-existent category: %d, want 404", res.StatusCode)
	}
	res.Body.Close()

	// Unlink with invalid product UUID → 400
	if res = ta.do(t, "POST", "/api/v1/products/not-a-uuid/categories/unlink",
		map[string]string{"category_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7"}, cookie); res.StatusCode != http.StatusBadRequest {
		t.Errorf("unlink bad product uuid: %d, want 400", res.StatusCode)
	}
	res.Body.Close()

	// Unlink with malformed JSON → 400
	if res = ta.do(t, "POST", "/api/v1/products/"+created.ID+"/categories/unlink",
		json.RawMessage("{oops"), cookie); res.StatusCode != http.StatusBadRequest {
		t.Errorf("unlink malformed: %d, want 400", res.StatusCode)
	}
	res.Body.Close()

	// Unlink with non-existent category → 404
	if res = ta.do(t, "POST", "/api/v1/products/"+created.ID+"/categories/unlink",
		map[string]string{"category_id": fakeCatID}, cookie); res.StatusCode != http.StatusNotFound {
		t.Errorf("unlink non-existent category: %d, want 404", res.StatusCode)
	}
	res.Body.Close()
}

func TestRoutes_ProductGetById_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "getbyid@example.com")

	// Create with categories
	catRes := ta.do(t, "POST", "/api/v1/categories", map[string]string{"name": "Electronics"}, cookie)
	var cat struct {
		ID string `json:"id"`
	}
	decodeBody(t, catRes, &cat)

	prodRes := ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "Laptop", "description": "gaming laptop", "price": 99999, "stock_quantity": 10,
	}, cookie)
	var prod struct {
		ID string `json:"id"`
	}
	decodeBody(t, prodRes, &prod)

	ta.do(t, "POST", "/api/v1/products/"+prod.ID+"/categories/link",
		map[string]string{"category_id": cat.ID}, cookie)

	// Get by ID → 200 with categories (ResponseProduct embeds Product at top level)
	var got struct {
		Name       string `json:"name"`
		Categories []struct {
			Name string `json:"name"`
		} `json:"categories"`
	}
	decodeBody(t, ta.do(t, "GET", "/api/v1/products/"+prod.ID, nil, ""), &got)
	if got.Name != "Laptop" {
		t.Errorf("get by id: name=%q, want Laptop", got.Name)
	}
	if len(got.Categories) != 1 || got.Categories[0].Name != "electronics" {
		t.Errorf("get by id: categories=%+v", got.Categories)
	}

	// Get by ID with bad uuid → 400
	if res := ta.do(t, "GET", "/api/v1/products/not-a-uuid", nil, ""); res.StatusCode != http.StatusBadRequest {
		t.Errorf("get by bad id: %d, want 400", res.StatusCode)
	} else {
		res.Body.Close()
	}

	// Get by ID not found → 404
	zeroID := "00000000-0000-0000-0000-000000000000"
	if res := ta.do(t, "GET", "/api/v1/products/"+zeroID, nil, ""); res.StatusCode != http.StatusNotFound {
		t.Errorf("get by non-existent id: %d, want 404", res.StatusCode)
	} else {
		res.Body.Close()
	}
}

func TestRoutes_ProductGetByName_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "getbyname@example.com")

	res := ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "Keyboard", "description": "mechanical", "price": 7999, "stock_quantity": 50,
	}, cookie)
	res.Body.Close()

	// Get by name → 200 (ResponseProduct embeds Product at top level)
	var got struct {
		Name string `json:"name"`
	}
	decodeBody(t, ta.do(t, "GET", "/api/v1/products/name/Keyboard", nil, ""), &got)
	if got.Name != "Keyboard" {
		t.Errorf("get by name: name=%q, want Keyboard", got.Name)
	}

	// Get by name not found → 404
	if res := ta.do(t, "GET", "/api/v1/products/name/NonExistent", nil, ""); res.StatusCode != http.StatusNotFound {
		t.Errorf("get by non-existent name: %d, want 404", res.StatusCode)
	} else {
		res.Body.Close()
	}
}

func TestRoutes_ProductUnlink_Success_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "unlink-ok@example.com")

	// Create category
	catRes := ta.do(t, "POST", "/api/v1/categories", map[string]string{"name": "Shoes"}, cookie)
	var cat struct {
		ID string `json:"id"`
	}
	decodeBody(t, catRes, &cat)

	// Create product
	prodRes := ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "Sneakers", "description": "comfy", "price": 5999, "stock_quantity": 20,
	}, cookie)
	var prod struct {
		ID string `json:"id"`
	}
	decodeBody(t, prodRes, &prod)

	// Link
	linkRes := ta.do(t, "POST", "/api/v1/products/"+prod.ID+"/categories/link",
		map[string]string{"category_id": cat.ID}, cookie)
	var linked struct {
		Categories []struct{ Name string } `json:"categories"`
	}
	decodeBody(t, linkRes, &linked)
	if len(linked.Categories) != 1 {
		t.Fatalf("after link: %d categories, want 1", len(linked.Categories))
	}

	// Unlink → 200 with empty categories
	unlinkRes := ta.do(t, "POST", "/api/v1/products/"+prod.ID+"/categories/unlink",
		map[string]string{"category_id": cat.ID}, cookie)
	decodeBody(t, unlinkRes, &linked)
	if len(linked.Categories) != 0 {
		t.Errorf("after unlink: %d categories, want 0", len(linked.Categories))
	}
}

func TestRoutes_ProductDelete_Success_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "del-ok@example.com")

	// Create product
	prodRes := ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "ToDelete", "description": "will be deleted", "price": 100, "stock_quantity": 1,
	}, cookie)
	var prod struct {
		ID string `json:"id"`
	}
	decodeBody(t, prodRes, &prod)

	// Delete → 204
	if res := ta.do(t, "DELETE", "/api/v1/products/"+prod.ID, nil, cookie); res.StatusCode != http.StatusNoContent {
		t.Errorf("delete: %d, want 204", res.StatusCode)
	} else {
		res.Body.Close()
	}

	// Verify gone → 404
	if res := ta.do(t, "GET", "/api/v1/products/"+prod.ID, nil, ""); res.StatusCode != http.StatusNotFound {
		t.Errorf("get after delete: %d, want 404", res.StatusCode)
	} else {
		res.Body.Close()
	}
}

func TestRoutes_CategoryList_DefaultLimit_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "cat-list-default@example.com")

	// Create 3 categories
	for _, name := range []string{"Alpha", "Beta", "Gamma"} {
		res := ta.do(t, "POST", "/api/v1/categories", map[string]string{"name": name}, cookie)
		res.Body.Close()
	}

	// List with default limit (no limit param → defaults to 20)
	var list struct {
		Categories []struct{ Name string } `json:"categories"`
		Total      int64                   `json:"total"`
		Limit      int                     `json:"limit"`
	}
	decodeBody(t, ta.do(t, "GET", "/api/v1/categories", nil, ""), &list)
	if list.Total != 3 {
		t.Errorf("default list total=%d, want 3", list.Total)
	}
	if list.Limit != 20 {
		t.Errorf("default list limit=%d, want 20", list.Limit)
	}
}

func TestRoutes_CategoryList_FilterByName_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "cat-filter-test@example.com")

	for _, name := range []string{"Electronics", "Electricals", "Clothing"} {
		res := ta.do(t, "POST", "/api/v1/categories", map[string]string{"name": name}, cookie)
		res.Body.Close()
	}

	var list struct {
		Categories []struct{ Name string } `json:"categories"`
		Total      int64                   `json:"total"`
	}

	// Filter by substring "elect"
	decodeBody(t, ta.do(t, "GET", "/api/v1/categories?name=elect", nil, ""), &list)
	if list.Total != 2 || len(list.Categories) != 2 {
		t.Fatalf("expected 2 categories matching 'elect', got total=%d len=%d", list.Total, len(list.Categories))
	}

	// Filter by non-matching
	decodeBody(t, ta.do(t, "GET", "/api/v1/categories?name=nonexistent", nil, ""), &list)
	if list.Total != 0 || len(list.Categories) != 0 {
		t.Fatalf("expected 0 categories, got total=%d len=%d", list.Total, len(list.Categories))
	}
}

func TestRoutes_ProductList_Filters_E2E(t *testing.T) {
	ta := newTestApp(t)
	cookie := ta.loginAs(t, "filter-test@example.com")

	// 1. Create categories
	var cat1, cat2 struct{ ID, Name string }
	decodeBody(t, ta.do(t, "POST", "/api/v1/categories", map[string]string{"name": "Gadgets"}, cookie), &cat1)
	decodeBody(t, ta.do(t, "POST", "/api/v1/categories", map[string]string{"name": "Apparel"}, cookie), &cat2)

	// 2. Create products
	var p1, p2, p3 struct{ ID, Name string }
	decodeBody(t, ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "Smart Watch", "description": "Gadget watch", "price": 100, "stock_quantity": 5,
	}, cookie), &p1)
	decodeBody(t, ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "Running Shorts", "description": "Sportswear shorts", "price": 50, "stock_quantity": 10,
	}, cookie), &p2)
	decodeBody(t, ta.do(t, "POST", "/api/v1/products", map[string]any{
		"name": "Smart Glasses", "description": "High tech glasses", "price": 200, "stock_quantity": 2,
	}, cookie), &p3)

	// Link categories: p1 -> Gadgets, p2 -> Apparel, p3 -> Gadgets & Apparel
	ta.do(t, "POST", "/api/v1/products/"+p1.ID+"/categories/link", map[string]string{"category_id": cat1.ID}, cookie).Body.Close()
	ta.do(t, "POST", "/api/v1/products/"+p2.ID+"/categories/link", map[string]string{"category_id": cat2.ID}, cookie).Body.Close()
	ta.do(t, "POST", "/api/v1/products/"+p3.ID+"/categories/link", map[string]string{"category_id": cat1.ID}, cookie).Body.Close()
	ta.do(t, "POST", "/api/v1/products/"+p3.ID+"/categories/link", map[string]string{"category_id": cat2.ID}, cookie).Body.Close()

	type productListResp struct {
		Products []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"products"`
		Total int64 `json:"total"`
	}

	// Test A: Name search substring "watch"
	var resA productListResp
	decodeBody(t, ta.do(t, "GET", "/api/v1/products?name=watch", nil, ""), &resA)
	if resA.Total != 1 || len(resA.Products) != 1 || resA.Products[0].Name != "Smart Watch" {
		t.Fatalf("expected Smart Watch, got total=%d, len=%d", resA.Total, len(resA.Products))
	}

	// Test B: Filter by single category (Apparel: p2, p3)
	var resB productListResp
	decodeBody(t, ta.do(t, "GET", "/api/v1/products?category_ids="+cat2.ID, nil, ""), &resB)
	if resB.Total != 2 || len(resB.Products) != 2 {
		t.Fatalf("expected 2 apparel products, got total=%d, len=%d", resB.Total, len(resB.Products))
	}

	// Test C: Filter by multiple categories comma-separated (Gadgets, Apparel: p1, p2, p3)
	var resC productListResp
	decodeBody(t, ta.do(t, "GET", "/api/v1/products?category_ids="+cat1.ID+","+cat2.ID, nil, ""), &resC)
	if resC.Total != 3 || len(resC.Products) != 3 {
		t.Fatalf("expected 3 products in Gadgets or Apparel, got total=%d, len=%d", resC.Total, len(resC.Products))
	}

	// Test D: Filter by multiple categories repeated query params
	var resD productListResp
	decodeBody(t, ta.do(t, "GET", "/api/v1/products?category_ids="+cat1.ID+"&category_ids="+cat2.ID, nil, ""), &resD)
	if resD.Total != 3 || len(resD.Products) != 3 {
		t.Fatalf("expected 3 products with repeated category_ids, got total=%d, len=%d", resD.Total, len(resD.Products))
	}

	// Test E: Name + Category filter (name "Smart" + category Apparel: should match only Smart Glasses)
	var resE productListResp
	decodeBody(t, ta.do(t, "GET", "/api/v1/products?name=smart&category_ids="+cat2.ID, nil, ""), &resE)
	if resE.Total != 1 || len(resE.Products) != 1 || resE.Products[0].Name != "Smart Glasses" {
		t.Fatalf("expected Smart Glasses, got total=%d, len=%d", resE.Total, len(resE.Products))
	}

	// Test F: My Products filter
	var resF productListResp
	decodeBody(t, ta.do(t, "GET", "/api/v1/my-products?name=shorts", nil, cookie), &resF)
	if resF.Total != 1 || len(resF.Products) != 1 || resF.Products[0].Name != "Running Shorts" {
		t.Fatalf("expected Running Shorts in my-products, got total=%d, len=%d", resF.Total, len(resF.Products))
	}
}
