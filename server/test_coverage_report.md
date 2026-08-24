# Go Test Coverage Report (Backend)

> [!NOTE]
> This report is generated from the Go test coverage profile using the `-coverpkg` flag, which attributes statements executed across package boundaries (such as repository or service calls made during E2E route tests) to their respective packages.

## Overall Statement Coverage: **90.2%**

### Package-Level Summary

| Package | Functions | Covered (>0%) | Avg Func Coverage |
| --- | --- | --- | --- |
| `src/controllers/v1` | 21 | 21 | 87.7% |
| `src/middleware` | 1 | 1 | 91.7% |
| `src/repositories` | 36 | 36 | 95.8% |
| `src/routes` | 1 | 1 | 100.0% |
| `src/services` | 38 | 38 | 95.7% |
| `utils` | 18 | 16 | 85.4% |

---

## Detailed Function Coverage

### Package `src/controllers/v1`
- **Total Functions**: 21
- **Covered Functions**: 21 (100.0%)
- **Average Function Coverage**: 87.7%

<details>
<summary>View functions in src/controllers/v1</summary>

| File | Line | Function | Coverage | Status |
| --- | --- | --- | --- | --- |
| [AuthController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/AuthController.go#L26) | 26 | `NewAuthController` | **100.0%** | 🟢 Fully Covered |
| [AuthController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/AuthController.go#L44) | 44 | `Register` | **100.0%** | 🟢 Fully Covered |
| [AuthController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/AuthController.go#L86) | 86 | `Login` | **100.0%** | 🟢 Fully Covered |
| [AuthController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/AuthController.go#L129) | 129 | `Logout` | **85.7%** | 🟡 Partially Covered |
| [AuthController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/AuthController.go#L154) | 154 | `Me` | **85.7%** | 🟡 Partially Covered |
| [AuthController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/AuthController.go#L174) | 174 | `setAuthCookies` | **100.0%** | 🟢 Fully Covered |
| [AuthController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/AuthController.go#L193) | 193 | `clearAuthCookies` | **100.0%** | 🟢 Fully Covered |
| [CategoryController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/CategoryController.go#L17) | 17 | `NewCategoryController` | **100.0%** | 🟢 Fully Covered |
| [CategoryController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/CategoryController.go#L35) | 35 | `MatchCategories` | **81.8%** | 🟡 Partially Covered |
| [CategoryController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/CategoryController.go#L77) | 77 | `ListCategories` | **90.9%** | 🟡 Partially Covered |
| [CategoryController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/CategoryController.go#L120) | 120 | `CreateCategory` | **84.6%** | 🟡 Partially Covered |
| [ProductController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/ProductController.go#L19) | 19 | `NewProductController` | **100.0%** | 🟢 Fully Covered |
| [ProductController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/ProductController.go#L37) | 37 | `ListProducts` | **90.9%** | 🟡 Partially Covered |
| [ProductController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/ProductController.go#L80) | 80 | `CreateProduct` | **90.9%** | 🟡 Partially Covered |
| [ProductController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/ProductController.go#L123) | 123 | `ListMyProducts` | **75.0%** | 🟡 Partially Covered |
| [ProductController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/ProductController.go#L166) | 166 | `GetProductById` | **100.0%** | 🟢 Fully Covered |
| [ProductController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/ProductController.go#L193) | 193 | `GetProductByName` | **100.0%** | 🟢 Fully Covered |
| [ProductController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/ProductController.go#L220) | 220 | `UpdateProduct` | **57.1%** | 🟡 Partially Covered |
| [ProductController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/ProductController.go#L272) | 272 | `LinkCategory` | **71.4%** | 🟡 Partially Covered |
| [ProductController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/ProductController.go#L324) | 324 | `UnlinkCategory` | **57.1%** | 🟡 Partially Covered |
| [ProductController.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/controllers/v1/ProductController.go#L372) | 372 | `DeleteProduct` | **71.4%** | 🟡 Partially Covered |

</details>

### Package `src/middleware`
- **Total Functions**: 1
- **Covered Functions**: 1 (100.0%)
- **Average Function Coverage**: 91.7%

<details>
<summary>View functions in src/middleware</summary>

| File | Line | Function | Coverage | Status |
| --- | --- | --- | --- | --- |
| [AuthMiddleware.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/middleware/AuthMiddleware.go#L11) | 11 | `RequireAuth` | **91.7%** | 🟡 Partially Covered |

</details>

### Package `src/repositories`
- **Total Functions**: 36
- **Covered Functions**: 36 (100.0%)
- **Average Function Coverage**: 95.8%

<details>
<summary>View functions in src/repositories</summary>

| File | Line | Function | Coverage | Status |
| --- | --- | --- | --- | --- |
| [CategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/CategoryRepository.go#L21) | 21 | `InitCategoryRepository` | **100.0%** | 🟢 Fully Covered |
| [CategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/CategoryRepository.go#L31) | 31 | `GetCategoryById` | **100.0%** | 🟢 Fully Covered |
| [CategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/CategoryRepository.go#L62) | 62 | `GetCategoryByIds` | **100.0%** | 🟢 Fully Covered |
| [CategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/CategoryRepository.go#L89) | 89 | `GetCategoryByNames` | **100.0%** | 🟢 Fully Covered |
| [CategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/CategoryRepository.go#L123) | 123 | `GetCategories` | **90.0%** | 🟡 Partially Covered |
| [CategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/CategoryRepository.go#L166) | 166 | `CreateCategory` | **87.5%** | 🟡 Partially Covered |
| [CategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/CategoryRepository.go#L196) | 196 | `MatchCategoriesByName` | **100.0%** | 🟢 Fully Covered |
| [CategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/CategoryRepository.go#L229) | 229 | `DeleteCategory` | **100.0%** | 🟢 Fully Covered |
| [OrderRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/OrderRepository.go#L21) | 21 | `InitOrderRepository` | **100.0%** | 🟢 Fully Covered |
| [OrderRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/OrderRepository.go#L30) | 30 | `ListOrders` | **90.0%** | 🟡 Partially Covered |
| [OrderRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/OrderRepository.go#L74) | 74 | `ListOrdersInRange` | **90.0%** | 🟡 Partially Covered |
| [OrderRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/OrderRepository.go#L123) | 123 | `DeleteOrder` | **100.0%** | 🟢 Fully Covered |
| [OrderRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/OrderRepository.go#L150) | 150 | `DeleteOrders` | **84.6%** | 🟡 Partially Covered |
| [ProductCategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductCategoryRepository.go#L19) | 19 | `InitProductCategoryRepository` | **100.0%** | 🟢 Fully Covered |
| [ProductCategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductCategoryRepository.go#L31) | 31 | `GetCategoriesByProduct` | **83.3%** | 🟡 Partially Covered |
| [ProductCategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductCategoryRepository.go#L62) | 62 | `GetCategoriesByProductIds` | **100.0%** | 🟢 Fully Covered |
| [ProductCategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductCategoryRepository.go#L95) | 95 | `GetProductCategory` | **100.0%** | 🟢 Fully Covered |
| [ProductCategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductCategoryRepository.go#L129) | 129 | `LinkCategory` | **92.3%** | 🟡 Partially Covered |
| [ProductCategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductCategoryRepository.go#L179) | 179 | `UnlinkCategory` | **100.0%** | 🟢 Fully Covered |
| [ProductCategoryRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductCategoryRepository.go#L206) | 206 | `DeleteProductCategories` | **100.0%** | 🟢 Fully Covered |
| [ProductRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductRepository.go#L20) | 20 | `InitProductRepository` | **100.0%** | 🟢 Fully Covered |
| [ProductRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductRepository.go#L28) | 28 | `GetProductById` | **100.0%** | 🟢 Fully Covered |
| [ProductRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductRepository.go#L63) | 63 | `GetProductByName` | **100.0%** | 🟢 Fully Covered |
| [ProductRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductRepository.go#L98) | 98 | `GetProducts` | **90.0%** | 🟡 Partially Covered |
| [ProductRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductRepository.go#L145) | 145 | `CreateProduct` | **91.7%** | 🟡 Partially Covered |
| [ProductRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductRepository.go#L177) | 177 | `UpdateProduct` | **100.0%** | 🟢 Fully Covered |
| [ProductRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductRepository.go#L202) | 202 | `DeleteProduct` | **100.0%** | 🟢 Fully Covered |
| [ProductRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/ProductRepository.go#L218) | 218 | `GetMyProducts` | **90.0%** | 🟡 Partially Covered |
| [UserRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/UserRepository.go#L21) | 21 | `InitUserRepository` | **100.0%** | 🟢 Fully Covered |
| [UserRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/UserRepository.go#L28) | 28 | `GetUserByEmail` | **100.0%** | 🟢 Fully Covered |
| [UserRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/UserRepository.go#L62) | 62 | `GetUserById` | **100.0%** | 🟢 Fully Covered |
| [UserRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/UserRepository.go#L96) | 96 | `CreateUser` | **87.5%** | 🟡 Partially Covered |
| [UserRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/UserRepository.go#L121) | 121 | `SoftDeleteUser` | **100.0%** | 🟢 Fully Covered |
| [UserRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/UserRepository.go#L136) | 136 | `CreateRefreshToken` | **83.3%** | 🟡 Partially Covered |
| [UserRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/UserRepository.go#L156) | 156 | `GetRefreshTokenByHash` | **87.5%** | 🟡 Partially Covered |
| [UserRepository.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/repositories/UserRepository.go#L187) | 187 | `DeleteRefreshTokensByUser` | **100.0%** | 🟢 Fully Covered |

</details>

### Package `src/routes`
- **Total Functions**: 1
- **Covered Functions**: 1 (100.0%)
- **Average Function Coverage**: 100.0%

<details>
<summary>View functions in src/routes</summary>

| File | Line | Function | Coverage | Status |
| --- | --- | --- | --- | --- |
| [v1.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/routes/v1.go#L9) | 9 | `RegisterV1Routes` | **100.0%** | 🟢 Fully Covered |

</details>

### Package `src/services`
- **Total Functions**: 38
- **Covered Functions**: 38 (100.0%)
- **Average Function Coverage**: 95.7%

<details>
<summary>View functions in src/services</summary>

| File | Line | Function | Coverage | Status |
| --- | --- | --- | --- | --- |
| [AuthService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/AuthService.go#L26) | 26 | `InitAuthService` | **100.0%** | 🟢 Fully Covered |
| [AuthService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/AuthService.go#L45) | 45 | `Register` | **95.2%** | 🟡 Partially Covered |
| [AuthService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/AuthService.go#L94) | 94 | `Login` | **92.3%** | 🟡 Partially Covered |
| [AuthService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/AuthService.go#L144) | 144 | `Logout` | **100.0%** | 🟢 Fully Covered |
| [CategoryService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/CategoryService.go#L16) | 16 | `InitCategoryService` | **100.0%** | 🟢 Fully Covered |
| [CategoryService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/CategoryService.go#L26) | 26 | `GetCategoryById` | **100.0%** | 🟢 Fully Covered |
| [CategoryService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/CategoryService.go#L30) | 30 | `GetCategoryByNames` | **100.0%** | 🟢 Fully Covered |
| [CategoryService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/CategoryService.go#L34) | 34 | `ListCategories` | **100.0%** | 🟢 Fully Covered |
| [CategoryService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/CategoryService.go#L50) | 50 | `MatchCategories` | **100.0%** | 🟢 Fully Covered |
| [CategoryService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/CategoryService.go#L57) | 57 | `CreateCategory` | **100.0%** | 🟢 Fully Covered |
| [CategoryService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/CategoryService.go#L85) | 85 | `DeleteCategory` | **100.0%** | 🟢 Fully Covered |
| [OrderService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/OrderService.go#L20) | 20 | `InitOrderService` | **100.0%** | 🟢 Fully Covered |
| [OrderService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/OrderService.go#L29) | 29 | `ListOrders` | **100.0%** | 🟢 Fully Covered |
| [OrderService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/OrderService.go#L46) | 46 | `ListOrdersInRange` | **100.0%** | 🟢 Fully Covered |
| [OrderService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/OrderService.go#L63) | 63 | `RemoveOrder` | **88.9%** | 🟡 Partially Covered |
| [OrderService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/OrderService.go#L83) | 83 | `RemoveOrders` | **100.0%** | 🟢 Fully Covered |
| [OrderService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/OrderService.go#L93) | 93 | `InTx` | **85.7%** | 🟡 Partially Covered |
| [OrderService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/OrderService.go#L111) | 111 | `ListOrdersInRangeTx` | **88.9%** | 🟡 Partially Covered |
| [OrderService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/OrderService.go#L130) | 130 | `RemoveOrdersTx` | **100.0%** | 🟢 Fully Covered |
| [ProductService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/ProductService.go#L25) | 25 | `InitProductService` | **100.0%** | 🟢 Fully Covered |
| [ProductService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/ProductService.go#L36) | 36 | `getProductsCategory` | **80.0%** | 🟡 Partially Covered |
| [ProductService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/ProductService.go#L56) | 56 | `GetProductById` | **100.0%** | 🟢 Fully Covered |
| [ProductService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/ProductService.go#L71) | 71 | `GetProductByName` | **100.0%** | 🟢 Fully Covered |
| [ProductService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/ProductService.go#L88) | 88 | `ListProducts` | **100.0%** | 🟢 Fully Covered |
| [ProductService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/ProductService.go#L152) | 152 | `CreateProduct` | **88.2%** | 🟡 Partially Covered |
| [ProductService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/ProductService.go#L190) | 190 | `ListMyProducts` | **91.4%** | 🟡 Partially Covered |
| [ProductService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/ProductService.go#L254) | 254 | `UpdateProduct` | **88.2%** | 🟡 Partially Covered |
| [ProductService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/ProductService.go#L292) | 292 | `LinkCategory` | **88.9%** | 🟡 Partially Covered |
| [ProductService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/ProductService.go#L328) | 328 | `UnlinkCategory` | **88.9%** | 🟡 Partially Covered |
| [ProductService.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/ProductService.go#L363) | 363 | `DeleteProduct` | **90.9%** | 🟡 Partially Covered |
| [dedup_service.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/dedup_service.go#L36) | 36 | `RunDedupOrdersRemove` | **80.4%** | 🟡 Partially Covered |
| [dedup_service.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/dedup_service.go#L154) | 154 | `ProcessDedupChunk` | **100.0%** | 🟢 Fully Covered |
| [dedup_service.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/dedup_service.go#L246) | 246 | `SplitRange` | **89.5%** | 🟡 Partially Covered |
| [dedup_service.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/dedup_service.go#L276) | 276 | `ClusterKeepLatest` | **100.0%** | 🟢 Fully Covered |
| [errors.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/errors.go#L23) | 23 | `IsDuplicateProductName` | **100.0%** | 🟢 Fully Covered |
| [errors.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/errors.go#L27) | 27 | `IsDuplicateCategoryName` | **100.0%** | 🟢 Fully Covered |
| [errors.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/errors.go#L31) | 31 | `IsDuplicateEmail` | **100.0%** | 🟢 Fully Covered |
| [errors.go](file:///Users/jayponda/drive/projects/product_catalog/server/src/services/errors.go#L35) | 35 | `isUniqueViolation` | **100.0%** | 🟢 Fully Covered |

</details>

### Package `utils`
- **Total Functions**: 18
- **Covered Functions**: 16 (88.9%)
- **Average Function Coverage**: 85.4%

<details>
<summary>View functions in utils</summary>

| File | Line | Function | Coverage | Status |
| --- | --- | --- | --- | --- |
| [db.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/db.go#L32) | 32 | `InitDB` | **0.0%** | 🔴 Uncovered |
| [db.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/db.go#L60) | 60 | `GetDB` | **0.0%** | 🔴 Uncovered |
| [executor.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/executor.go#L14) | 14 | `ResolveExecutor` | **100.0%** | 🟢 Fully Covered |
| [jwt.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/jwt.go#L14) | 14 | `GenerateSecureToken` | **75.0%** | 🟡 Partially Covered |
| [jwt.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/jwt.go#L23) | 23 | `HashToken` | **100.0%** | 🟢 Fully Covered |
| [jwt.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/jwt.go#L39) | 39 | `GenerateAccessToken` | **100.0%** | 🟢 Fully Covered |
| [jwt.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/jwt.go#L55) | 55 | `ParseAccessToken` | **90.0%** | 🟡 Partially Covered |
| [jwt.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/jwt.go#L74) | 74 | `GenerateRefreshToken` | **80.0%** | 🟡 Partially Covered |
| [logger.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/logger.go#L37) | 37 | `NewStructuredLogger` | **100.0%** | 🟢 Fully Covered |
| [logger.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/logger.go#L51) | 51 | `Debug` | **100.0%** | 🟢 Fully Covered |
| [logger.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/logger.go#L55) | 55 | `Info` | **100.0%** | 🟢 Fully Covered |
| [logger.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/logger.go#L59) | 59 | `Warn` | **100.0%** | 🟢 Fully Covered |
| [logger.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/logger.go#L63) | 63 | `Error` | **100.0%** | 🟢 Fully Covered |
| [logger.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/logger.go#L68) | 68 | `logWithLevel` | **92.3%** | 🟡 Partially Covered |
| [logger.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/logger.go#L99) | 99 | `getLogger` | **100.0%** | 🟢 Fully Covered |
| [normalize.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/normalize.go#L5) | 5 | `NormalizeName` | **100.0%** | 🟢 Fully Covered |
| [uuid.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/uuid.go#L5) | 5 | `GetUUID` | **100.0%** | 🟢 Fully Covered |
| [validator.go](file:///Users/jayponda/drive/projects/product_catalog/server/utils/validator.go#L10) | 10 | `NewValidator` | **100.0%** | 🟢 Fully Covered |

</details>
