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
