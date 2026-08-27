# Product-catalog
- this project is simple and can be use as a boilerplat for future projects.

# index

1. developer setup
    - a. postgres
        - i. docker
    - b. backend
        - i. golang and fiber
        - ii. air (watcher)
        - iii. migration (before server run)
        - iv. server (app)
    - c. frontend setup
        - i. node and vue3
        - ii. server (watcher)
    - d. git hooks setup
    - e. environment variables
2. Tools information
    - a. Backend
        - i. migration
        - ii. orm
        - iii. uuid
        - iv. testing
        - v. jwt + refresh token pattern (auth)
        - vi. cli (cobra and command pattern)
        - vii. swagger (api docs + visual testing)
    - b. Frontend
        - i. tailwind css
        - ii. testing
    - c. General
        - i. hooks
        - ii. command quick reference
3. production setup
    - a. config envs
    - b. docker setup
4. scope description
    - a. form simulator
    - b. duplicate order aggregator (Cli)
5. assumptions
6. project description
7. project structure
    - a. backend
    - b. frontend


## 1. Developer setup

- This project will be monorepo.
- So frontend will be in app and backend will be in server directory.
- Thired we use postgres as database and run inside the docker.

> <b>Note</b>:  we consider currunt directory as a root. and all paths are given here is releted to the root directory. 

### a. Postgres setup
- If you don't have a docker please install it.
- Second we have docker-compose.storage.yml file for database. 

```
   $ root > docker compose -f docker-compose.storage.yml up -d
```

### b. Backend setup
- the backend server is written in golang and we choose the framework fiber for it. 
- we choose the ORM goqu here. 
- go uses the binary to run, and in development it's not possible to rerun on change so we will use air for a filewatcher, it will help us in development mode, so we should not each time build and run the binary by our self.

#### i. golang and fiber
- you need to download golang for you local system.
- then go to server directory and download the dependancies.
```
$ root/server > go get
```

- this will download the fiber and other required package.



#### ii. air (watcher)
- download the golang air for watcher. 
- we have air.tomb file for the configuration. 
```
$ root > go install github.com/air-verse/air@latest
```

- then you need to add golang binary to your path.
```
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

- start an watcher
```
$ root/server > air
```

#### iii. migration (before server run)
- Database migrations are managed with [rubenv/sql-migrate](https://github.com/rubenv/sql-migrate), driven by a Cobra-based CLI defined in `server/commands/cmd`. The CLI exposes two top-level commands: `server` (starts the HTTP server) and `migrate` (migration tasks).
- Day-to-day commands are wrapped in Makefile targets:

1. `make mfile <name>` — generate a new migration file (scaffolds `-- +migrate Up` / `-- +migrate Down`)
2. `make mup` — apply all pending migrations
3. `make mdown [s=N]` — roll back the last migration (`s` = number of steps)
4. `make mseed` — one-time: register already-applied migrations after switching tools

```
$ root/server > make mfile create_category_table
$ root/server > make mup
$ root/server > make mdown
$ root/server > make mseed
```

- The same tasks are also available directly through the CLI:

```
$ root/server > go run . migrate up
$ root/server > go run . migrate down --steps 1
$ root/server > go run . migrate create create_category_table
$ root/server > go run . server
```

> note: always run `migrate up` before starting the server.
> note: migrations should run per deployment, not per server start.
> note: ensure envs (especially `DB_DIALECT` / `DB_*`) are configured.

#### iv. server (app)
- to start dev server, you need to do.
```
$ root/server > make dev
```

### c. Frontend setup

#### i. node and vue3

- we use node environment. here so please install the node.
- we use pnpm as a package manager so install that as well. 
> npx get-pnpm

- app directory have the code for frontend.
- first of all you need to install the dependacyes of the project.
```
$ root/app > pnpm install 
```

#### ii. server (watcher)

- you are ready to run this project. for that run below command. 
```
$ root/app > pnpm dev 
```

### d. Git hooks setup

To automate formatting and linting verification on staged files, run the following setup command from the repository root:

```bash
$ root > make setup-hooks
```

This target configures git to use the hooks stored in the `.githooks/` directory and ensures that the pre-commit script is executable.

The pre-commit hook runs the following checks based on the staged files:
- **Go files**: Runs `gofmt` to verify formatting and `golangci-lint` to check code quality.
- **Frontend files (JS, TS, Vue, CSS, JSON)**: Runs `prettier --check` to verify formatting, and `oxlint`/`eslint` to check code quality.

If any check fails, the commit will be blocked and the specific errors will be displayed. You can bypass the checks using `git commit --no-verify` if needed.

### e. Environment variables

Both apps read config from a `.env` file in their own directory — copy each `.env.example` to `.env` first. Never commit real secrets; only the `.example` files are tracked.

Server (`server/.env`) — loaded by godotenv, parsed via struct tags (`EnvConfig` in `server/main.go`). Required vars make startup fail fast:

| Variable | Required | Default | Purpose |
|---|---|---|---|
| `APP_ENV` | no | `local` | runtime env; anything other than `local` enables cookie `Secure` flag |
| `HOST` / `PORT` | no | `localhost` / `8080` | HTTP bind address (example file uses `3300` — keep in sync with frontend) |
| `ALLOWED_ORIGINS` | no | `http://localhost:3000,http://localhost:5173` | comma-separated CORS origins |
| `JWT_SECRET` | **yes** | — | HS256 signing secret; use a long random string |
| `ACCESS_TOKEN_TTL` | no | `15m` | access token lifetime (Go duration) |
| `REFRESH_TOKEN_TTL` | no | `168h` | refresh token lifetime (7 days) |
| `DB_HOST` `DB_PORT` `DB_USER` `DB_PASSWORD` `DB_NAME` | **yes** | — | postgres connection segments (see 1.a for the container) |
| `DB_SSLMODE` | no | `disable` | postgres sslmode |
| `DB_DIALECT` | no | `postgres` | SQL dialect used by goqu + sql-migrate |
| `DB_MAX_OPEN_CONNS` | no | `25` | pool: max open connections |
| `DB_MAX_IDLE_CONNS` | no | `25` | pool: max idle connections |
| `DB_MAX_LIFETIME` | no | `5m` | pool: connection max lifetime |
| `DB_MAX_IDLE_TIME` | no | `2m` | pool: idle timeout |

Frontend (`app/.env`) — Vite only exposes vars prefixed with `VITE_`:

| Variable | Required | Default | Purpose |
|---|---|---|---|
| `VITE_APP_ENV` | no | `local` | runtime env label |
| `VITE_BACKEND_URL` | no | `http://localhost:3300/` | base URL for all API calls; must match server `HOST`/`PORT` |

> note: durations follow Go's `time.Duration` format (`15m`, `6h`, `168h`).
> note: changing `PORT` on the server means updating `VITE_BACKEND_URL` in the app too.

## 2. Tools information

### a. Backend

#### i. Migration — rubenv/sql-migrate
- **Library:** [github.com/rubenv/sql-migrate](https://github.com/rubenv/sql-migrate)
- **CLI:** Cobra-based commands in `server/commands/cmd` (no global `migrate` binary required).
  - `server` — starts the HTTP server (wires env, logger, repositories, services and routes).
  - `migrate up` — apply all pending migrations.
  - `migrate down [--steps N]` — roll back `N` migrations (default 1).
  - `migrate create <name>` — scaffold a new migration file.
  - `migrate seed` — one-time: mark already-applied migrations (used when switching from the previous tool).
- **Migration format:** one file per migration containing `-- +migrate Up` and `-- +migrate Down` sections (see `server/migrations/`).
- **Dialect:** taken from the `DB_DIALECT` env (defaults to `postgres`).
- **Why this tool:** each migration runs inside a transaction and is only recorded after success, so a failing file is rolled back and never marked applied. Re-running `migrate up` simply retries it — there is no `dirty` state and no manual `force` step.

#### ii. ORM — doug-martin/goqu
- **Library:** [github.com/doug-martin/goqu/v9](https://github.com/doug-martin/goqu/v9)
- **Where:** all database access lives in `server/src/repositories` — every repository receives a `*goqu.Database`; controllers and services never build SQL themselves.
- **Connection:** created once via `utils.InitDB(cfg)` (`server/utils/db.go`) with the dialect from `DB_DIALECT` (default `postgres`, driver `lib/pq`) and pooled per env settings (`DB_MAX_OPEN_CONNS`, etc.). Inside requests it is fetched with `utils.GetDB(ctx)` from the fiber context.
- **Style:** goqu is an SQL *builder*, not a full ORM — queries stay explicit: `db.From(...).Where(goqu.C("deleted_at").IsNull())...`, inserts/updates use `goqu.Record{}`, raw expressions fall back to `goqu.L(...)`.
- **Soft delete:** there are no ORM hooks; every read adds `deleted_at IS NULL` explicitly (see any repository) and deletes write `deleted_at = now()`.
- **Transactions:** multi-repository writes go through the Executor pattern (`utils.ResolveExecutor`, `server/utils/executor.go`) so services can run on either a plain DB or an open transaction — commit/rollback stays in one place.

#### iii. UUID — google/uuid
- **Library:** [github.com/google/uuid](https://github.com/google/uuid)
- **Where:** wrapped in `server/utils/uuid.go` (`utils.GetUUID()`); used for entity IDs instead of auto-increment integers.
- **Why this tool:** UUIDs avoid exposing record counts through enumerable IDs in the API and make it safe to generate IDs in application code before insert.

#### iv. Testing & coverage

Run everything from the repo root via the root Makefile:

```
$ root > make test            # backend + frontend suites
$ root > make coverage        # both coverage reports
```

Backend (`server/`):

| Command | What it does |
|---|---|
| `make test` | runs all Go unit + E2E tests |
| `make coverage` | same, plus a function-level coverage table |
| `make cover-html` | generates `cover.html` (per-line browser report) |

**Strategy — real database, zero mocks where it matters.**
Tests run against real in-memory SQLite databases ([modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite), pure Go, no CGO) with a schema mirrored from the production Postgres migrations, including the partial unique indexes (`uq_products_name_active`, `uq_categories_name_active`, `uq_users_email_active`). The helper lives in `server/testdb/testdb.go`:

- `testdb.OpenSQLite(t)` — isolated per-test DB, schema applied, auto-closed
- `testdb.SeedOrder` / `SeedProduct` — deterministic fixture inserts

Layered bottom-up:

1. **utils** — logger, JWT round-trips/expiry/wrong-secret, UUID, name normalization, validator rules
2. **repositories** — CRUD, soft-delete visibility, pagination ordering, refresh-token lifecycle, tombstone reactivation on re-link; plus `go-sqlmock` unit tests for query shapes
3. **services** — transactions verified end-to-end (commit *and* rollback paths), duplicate-name pre-check + constraint races, category hydration, auth (bcrypt, token issuance/revocation)
4. **routes** (E2E) — the whole stack over HTTP via fiber's `app.Test`: registration → login cookies → guarded product CRUD → link/unlink → logout cookie clearing; status-code contract asserted per route
5. **commands/cmd** (`dedup-orders-remove`) — pure-function tests for clustering/splitting, E2E runs of the real pipeline over SQLite (keep-latest verified in the DB, dry-run leaves rows intact), and a fault-injecting store stub proving each chunk's scan+delete is all-or-nothing: a mid-scan or mid-delete failure rolls back before anything is applied. Chunks are independent and reruns are idempotent (verified by executing the full run twice)

> note: the mirrored SQLite schema requires time columns written as `time.Time` values and non-NULL text for plain-Go model fields — see comments in `testdb/testdb.go`.

#### v. JWT + refresh token pattern (auth)

- **Libraries:** [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt/v5) for access tokens, stdlib `crypto/rand` + SHA-256 for refresh tokens, bcrypt for passwords — helpers live in `server/utils/jwt.go`, flow in `server/src/services/AuthService.go`.
- **Access token:** HS256-signed JWT (`utils.GenerateAccessToken`) carrying the user id plus standard `iat`/`exp` claims; TTL from `ACCESS_TOKEN_TTL` (default 15m). Parsing (`utils.ParseAccessToken`) rejects unexpected signing methods and expired tokens.
- **Refresh token:** 32 cryptographically random bytes, hex-encoded (`utils.GenerateSecureToken`). Only its SHA-256 hash is persisted in the `refresh_tokens` table with an expiry (`REFRESH_TOKEN_TTL`, default 168h = 7 days) — the raw value exists only in the client's cookie.
- **Delivery:** on register/login the controller sets both as httpOnly cookies (`access_token`, `refresh_token`). `SecureCookies` is enabled whenever `APP_ENV != "local"`, so production cookies require HTTPS.
- **Guard:** `middleware.RequireAuth(secret)` reads the `access_token` cookie, validates it and stores the parsed user UUID in `ctx.Locals(utils.UserContextKey)` for controllers.
- **Logout:** revokes *all* of the user's refresh tokens server-side (effective logout everywhere) and clears both cookies.
- note: there is no `/auth/refresh` endpoint yet — a fresh pair is issued at every login. The stored hashes still allow server-side session revocation.

#### vi. CLI — Cobra & command pattern

- **Library:** [spf13/cobra](https://github.com/spf13/cobra); commands in `server/commands/cmd/`.
- **One binary, many jobs:** `go run . <command>` — `server` (HTTP server), `migrate up|down|create|seed`, `dedup-orders-remove`.
- **Dependency injection over globals:** `main.go` parses the env config once and calls `cmd.Execute(cfg, logger)`, which stashes both into package vars (`root.go`). Commands build their repositories/services from that injected config instead of re-reading the environment.
- **Contract:** the `AppConfig` interface (`commands/cmd/root.go`) embeds `utils.DBConfigProvider` and adds host/port/CORS/JWT getters — anything the CLI needs must be declared there.
- **Adding a new command** (pattern documented in `root.go` comments):
  1. create a file in `server/commands/cmd/`
  2. define a `cobra.Command` with `Use`, flags and a `RunE`
  3. register it via `init()` → `rootCmd.AddCommand(...)`
  4. consume the shared `appConfig` / `appLogger`

#### vii. Swagger — API docs & visual testing

- **Libraries:** [swaggo/swag](https://github.com/swaggo/swag) annotations on controllers + `gofiber/contrib/v3/swaggo` middleware.
- **Generation:** `$ root/server > make swagger` scans from `main.go` and writes `server/docs/` (`docs.go`, `swagger.json`, `swagger.yaml`). Regenerate after changing handlers/models or the served docs drift.
- **Visual testing:** while the server runs, the docs are browsable and executable at `GET /docs/*` (BasePath `/api/v1`, wired in `serve.go`) — open `http://localhost:<PORT>/docs/` in a browser to inspect schemas and fire requests against the running API without curl.

#### viii. Other backend libraries (quick reference)

| Library | Purpose |
|---|---|
| `gofiber/fiber/v3` | HTTP framework: routing, middleware, cookies, and E2E testing via `app.Test` |
| `caarlos0/env/v11` + `joho/godotenv` | `.env` loading and struct-tag based config parsing (`server/main.go`) |
| `go-playground/validator/v10` | request payload validation rules (`server/utils/validator.go`) |
| `golang.org/x/crypto/bcrypt` | password hashing |
| `air` | dev file watcher (`make dev`), config in `server/.air.toml` |
| `golangci-lint` | linter (`make lint`), config in `server/.golangci.yml` |

### b. Frontend

Stack: Vue 3 (`<script setup>`) + Vite + pnpm.

#### i. UI toolkit — Tailwind CSS + Lucide icons
- **Styling:** [Tailwind CSS v4](https://tailwindcss.com) wired through `@tailwindcss/vite`; utilities used directly in SFCs, theme in `app/tailwind.config.js`.
- **Icons:** [`@lucide/vue`](https://lucide.dev) component set.
- **State & routing:** Pinia stores (`app/src/stores/auth.ts`) and `vue-router` (`app/src/router`) with auth guards on protected routes.
- **HTTP:** thin `fetch` wrapper in `app/src/network/request.js` returning a `{ ok, data | error }` result shape.
- **Dev server / build:** Vite (`pnpm dev` / `pnpm build`), Vue DevTools plugin enabled in dev.

#### ii. Testing & coverage

Frontend (`app/`):

| Command | What it does |
|---|---|
| `pnpm test:unit` | Vitest in watch mode |
| `pnpm test:unit -- --run` | single pass (CI style) |
| `pnpm coverage` | V8 coverage, text + HTML (`coverage/index.html`) |

Tooling quick reference:

| Tool | Purpose |
|---|---|
| Vitest + jsdom | unit/component tests |
| `@vue/test-utils` | component mounting and interaction |
| `@vitest/coverage-v8` | coverage reports |
| Prettier | formatting (`pnpm format`), config in `app/.prettierrc.json` |
| ESLint + oxlint | linting (`pnpm lint`); oxlint runs first for speed, ESLint covers the remaining rules incl. Vue |

Layers covered:

1. **network** (`src/network/request.js`) — URL/method/body/credentials for every endpoint wrapper, `{ ok, data | error }` result shape on success, HTTP error statuses, network failures, and error-message extraction for auth calls
2. **stores** (`src/stores/auth.ts`) — login/register/logout/fetchMe state transitions including failure paths leaving state untouched
3. **components** — `Header.vue`: logged-in vs logged-out UI, logout clearing session + navigating, mobile menu toggle
4. **router guard** — protected-route redirects (`/my-products`, `/products/add`, `/categories/add`, `/products/:id/edit`) carry a `redirect` query; public routes stay open; session restored exactly once

### c. General

#### i. Hooks

Git hooks live in `scripts/githooks/` and are wired with `make setup-hooks` (see section 1.d). Each hook is a thin shell script that runs a Node "engine" (`scripts/githooks/engine.js`) which resolves the changed files, loads plugins from `scripts/githooks/plugins/`, and executes tasks as a dependency chain — a task only runs if its `dependsOn` tasks passed, otherwise it is skipped as blocked. Full output of every run is written to `.githooks/tmp/<hook-name>.log` (kept out of git via `.gitignore`).

All available hook plugins:

| Plugin (file in `scripts/githooks/plugins/`) | What it does | Utilised in |
|---|---|---|
| `no-direct-main.js` | Blocks commits made directly on `main`/`master`; forces work on a feature branch merged via PR | pre-commit |
| `branch-name.js` | Branch must start with `feature/`, `hotfix/`, `bugfix/`, `release/`, `chore/` or `refactor/` (`main`/`master`/detached HEAD pass) | pre-commit |
| `gofmt.js` | Fails when changed `.go` files are not gofmt-formatted (check only — run `make format-backend` to fix) | pre-commit, pre-push |
| `golangci-lint.js` | Runs golangci-lint v2 over the whole `server/` tree (config: `server/.golangci.yml`) | pre-commit, pre-push |
| `prettier.js` | Runs `prettier --check` on changed frontend files (run `pnpm --dir app run format` to fix) | pre-commit, pre-push |
| `eslint-oxlint.js` | Lints changed frontend files: oxlint first (fast), then ESLint (incl. Vue rules) | pre-commit, pre-push |
| `go-coverage.js` | Runs the full backend test suite with `-coverpkg` and fails below **90%** statement coverage | pre-push |
| `vitest-coverage.js` | Runs Vitest with V8 coverage and fails below **90%** statement threshold | pre-push |
| `tag-validator.js` | Validates pushed tags: must be `vX.Y.Z` semver, point to a commit on `main`, be strictly greater than the highest existing tag, and not duplicate a tag on a different commit | pre-push |
| `sync-envs.js` | Runs on every push regardless of changed files: scans the repo for `.env` files and requires each to sit beside a `.env.example` with an identical key set (values are unrestricted; drift is caught in both directions — keys only in `.env` *and* keys only in the example). An example without a local `.env` (fresh clone) is skipped. It cannot rely on file matching because gitignored `.env` files never appear in changed-file lists | pre-push |
| `commit-msg-format.js` | Enforces Conventional Commits (`feat\|fix\|chore\|docs\|style\|refactor\|perf\|test(scope)?: message`); skips Merge/Revert/squash!/fixup! messages | commit-msg |
| `go-test.js` | Plain `go test ./...` without coverage — kept as an alternative to `go-coverage` | available (not registered) |
| `vitest.js` | Plain Vitest run without coverage — kept as an alternative to `vitest-coverage` | available (not registered) |

Plugins are small single-purpose modules (`no-direct-main`, `branch-name`, `gofmt`, `golangci-lint`, `go-coverage`, `prettier`, `eslint-oxlint`, `vitest-coverage`, `tag-validator`, `sync-envs`, `commit-msg-format`). Adding a new check means dropping a plugin file in `plugins/` and registering it in the relevant hook file — no changes to the engine needed. Bypass anything with `git commit --no-verify` when required.

#### ii. Command quick reference

Root Makefile — run from the **repo root**:

| Command | What it does |
|---|---|
| `make test` | runs backend + frontend test suites |
| `make coverage` | coverage reports for both sides |
| `make lint` | lints backend (golangci-lint) + frontend (oxlint/eslint) |
| `make format` | formats backend (`go fmt`) + frontend (prettier) |
| `make setup-hooks` | wires git hooks from `.githooks/` (see 1.d / 2.c.i) |

Server Makefile — run from **`server/`** (loads `.env` automatically):

| Command | What it does |
|---|---|
| `make dev` | starts dev server with air file watcher |
| `make prod` | regenerates Swagger docs, builds `bin/server`, runs it |
| `make swagger` | regenerates OpenAPI docs into `docs/` |
| `make mfile <name>` | scaffolds a new migration file |
| `make mup` | applies all pending migrations |
| `make mdown [s=N]` | rolls back N migrations (default 1) |
| `make mseed` | one-time: registers already-applied migrations after a tool switch |
| `make dedup ARGS="..."` | duplicate-order removal CLI; pass flags via `ARGS` (see 4.b) |
| `make dedup-dry-range START=... END=...` | dedup dry-run over an explicit RFC3339 range |
| `make test` | all Go unit + E2E tests |
| `make coverage` | tests + function-level coverage table |
| `make cover-html` | generates and opens `cover.html` per-line report |
| `make format` | `go fmt ./...` |
| `make lint` | golangci-lint over the server tree |

Frontend — run from **`app/`** (no makefile; pnpm scripts):

| Command | What it does |
|---|---|
| `pnpm dev` | Vite dev server with HMR |
| `pnpm build` | production build to `dist/` |
| `pnpm preview` | serves the production build locally |
| `pnpm test:unit [-- --run]` | Vitest watch mode (or single CI-style pass) |
| `pnpm coverage` | Vitest + V8 coverage report |
| `pnpm lint` | oxlint then eslint, auto-fixing where possible |
| `pnpm format` | prettier write over `src/` |


## 3. Production setup

The app ships as two container images (API + SPA behind nginx) orchestrated by `docker-compose.prod.yml`. Migrations run as a one-shot job before the API starts, and the database is never reachable from the frontend container.

### a. Config envs

All runtime configuration is injected as environment variables — no `.env` file is baked into any image (`godotenv` simply falls through to system env when `.env` is absent).

Secrets live in an env file next to the compose file, created from the template:

```
$ root > cp deploy/.env.production.example deploy/.env.production   # fill in real values; NEVER commit it
$ root > docker compose -f docker-compose.prod.yml --env-file deploy/.env.production up -d --build
```

Production-specific behaviour worth knowing:

| Variable / setting | Production notes |
|---|---|
| `APP_ENV=prod` | enables cookie `Secure` flag (auth requires HTTPS) and **disables Swagger** `/docs/*` |
| `ALLOWED_ORIGINS` | exact deployed frontend origin(s), comma separated, no trailing slash |
| `JWT_SECRET` | long random string — generate with `openssl rand -base64 48`; rotating it revokes every session |
| `DB_SSLMODE` | `require` minimum, `verify-full` when a CA bundle is available |
| `PUBLIC_API_URL` | leave **empty** for same-origin deployments (nginx proxies `/api`); set a full URL only for cross-origin hosting |
| `HTTP_PORT` | host port published for the web container (only public port of the whole stack) |

### b. Docker setup

**Images**

| Image | Build | Runtime |
|---|---|---|
| API (`server/Dockerfile`) | multi-stage: Go 1.26 build with vendored deps, `-trimpath -ldflags="-s -w"`, CGO off | alpine, non-root user, `ca-certificates` for TLS DBs; `migrations/` copied in (sql-migrate reads them relative to the workdir). Entrypoint runs `server` by default, `migrate up` via command override |
| Web (`app/Dockerfile`) | multi-stage: node 22 + pnpm `--frozen-lockfile`, `VITE_BACKEND_URL` injected as build arg | nginx alpine serving `dist/`; config in `app/docker/nginx.conf` |

**nginx responsibilities** (`app/docker/nginx.conf`):
- SPA fallback — deep links like `/products/123` resolve to `index.html` on hard refresh instead of 404ing
- same-origin reverse proxy — `/api/*` → `api:3300`, so there is no CORS surface and auth cookies stay first-party
- security headers (`nosniff`, `DENY`, referrer policy), gzip, immutable caching of hashed assets while `index.html` is never cached

**Orchestration flow** (`docker-compose.prod.yml`):

1. `postgres` becomes healthy (self-hosted option; use managed Postgres in real production)
2. `migrate` one-shot job applies pending migrations — if it exits non-zero, nothing else starts
3. `api` starts only after migrate completes successfully; turns healthy once `GET /readyz` (DB ping) passes
4. `web` starts only after api is healthy and publishes the single public port

**Health & shutdown:** the API exposes `GET /health` (liveness — process up) and `GET /readyz` (readiness — dependencies reachable), so platforms can route traffic correctly without killing healthy-but-slow containers. SIGTERM triggers a graceful drain of in-flight requests (25 s budget) before exit.

**Security hardening applied:** non-root containers, read-only root filesystems with tmpfs scratch space, `cap_drop: ALL` (web adds back only the five capabilities nginx needs), `no-new-privileges`, secrets passed at runtime only (`.dockerignore` keeps every `.env*` out of build contexts), pinned image versions, and network segmentation — `frontend` (web ↔ api) vs `backend` (`internal: true`, api/migrate/postgres, no internet egress). A compromised web container cannot resolve, let alone reach, the database.

**Rollback:** images are rebuilt from source each deploy; redeploy the previous git commit to roll back. Migrations are forward-only by design — write schema changes to be backward-compatible for one release (expand → deploy → contract).


## 4. Scope description

### a. Form simulator

Implemented — this is the form-driven catalog management flow in the frontend, exercising the product CRUD and category link/unlink endpoints end-to-end:

1. **Create / edit product** — `app/src/components/Product/ProductForm.vue` is one shared form used by both the create page (`/products/add`) and the edit page (`/products/:id/edit`, wrapped by `views/ModifyProduct.vue`).
   - Fields: name, description, stock quantity, price.
   - Client-side validation mirrors backend rules: required values, length caps (name ≤ 50, description ≤ 150), text must contain at least one letter, invisible/non-renderable characters rejected, integer stock ≥ 1 (≤ int32 max), price converted from dollars to cents (> 0, ≤ $9,999,999.99).
   - Create submits `POST /api/v1/products`; edit pre-fills the form from `GET /api/v1/products/:id` and saves via `PUT /api/v1/products/:id`.
2. **Link / unlink categories** — `app/src/components/Product/CategoryLinker.vue` on the edit page.
   - Already-linked categories render as tags; the ✕ button unlinks immediately (`POST /api/v1/products/:id/categories/unlink`).
   - Category search hits `/api/v1/categories/match` with a 250ms debounce; picked categories queue as pending tags and are linked in one go via `POST /api/v1/products/:id/categories/link`.
   - Per-category failures during linking are collected and surfaced through the shared error banner.

### b. Duplicate order aggregator (Cli)

Implemented as the `dedup-orders-remove` Cobra command (`server/commands/cmd/dedup_orders_remove.go`, logic in `server/src/services/dedup_service.go`). It cleans up accidental duplicate order submissions: orders created within a small gap of each other are clustered and only the **latest** of each cluster survives.

**Two run modes:**

1. **Scheduled/auto mode** — looks back over a sliding window ending at *now*. Default window is 2.5h (a 2h cron interval plus 0.5h overlap, so clusters straddling a boundary are still caught).
2. **Manual mode** — pass an explicit RFC3339 range with `--starttime`/`--endtime` (both required, end must be after start).

```
$ root/server > make dedup ARGS="--dry-run"
$ root/server > make dedup ARGS="--starttime 2024-01-01T00:00:00Z --endtime 2024-01-01T03:00:00Z"
$ root/server > make dedup ARGS="--window 2h30m --nearby 5m --batch 200"

# dry-run over an explicit range:
$ root/server > make dedup-dry-range START="2024-01-01T00:00:00Z" END="2024-01-01T03:00:00Z"
```

| Flag | Default | Meaning |
|---|---|---|
| `--dry-run` | off | report duplicates without deleting |
| `--nearby` | `5m` | max gap between two orders' `created_at` to treat them as duplicates |
| `--window` | `150m` | auto-mode lookback window |
| `--starttime` / `--endtime` | — | manual-mode RFC3339 bounds (both required) |
| `--batch` | `200` | fetch batch size per chunk |

**How it works:**

1. The requested range is split into overlapping chunks (`SplitRange`) so a cluster spanning a chunk boundary is never missed.
2. Per chunk, orders are fetched sorted by `created_at`; consecutive rows whose gap is ≤ `nearby` form a cluster (`ClusterKeepLatest`) and every order except the cluster's last is removed.
3. Removal is all-or-nothing per chunk: if scanning or deleting fails mid-way, nothing from that chunk is applied.
4. Reruns are idempotent — after a successful pass, a second identical run finds nothing to remove.
5. With `--dry-run`, clusters are reported but no rows are touched.

The behaviour is covered by tests: clustering/splitting pure functions, E2E runs over SQLite (keep-latest verified in the DB), dry-run leaving rows intact, and a fault-injecting store proving chunk-level atomicity (see section 2.a.iv).


## 5. Assumptions

1. **Monorepo** — frontend (`app/`) and backend (`server/`) live in one repo; all doc paths are relative to the repo root.
2. **Postgres in dev and prod** (via `docker-compose.storage.yml`); SQLite exists only inside the test suite as a mirrored schema.
3. **Soft delete everywhere** — rows carry `deleted_at`; "active" means `deleted_at IS NULL`. Uniqueness is enforced by partial unique indexes (`uq_*_active`) so tombstoned names can be reused.
4. **UUID primary keys** generated in application code (`utils.GetUUID()`), not by the database.
5. **Auth model** — short-lived JWT access tokens (~15m) plus rotating refresh tokens stored hashed server-side; both delivered as httpOnly cookies. `APP_ENV=prod` switches cookies to `Secure`.
6. **Migrations run per deployment**, never on server boot; the app assumes the schema it finds is current.
7. **One binary, many jobs** — the same CLI serves `server`, migration tasks and maintenance commands (dedup); ops tooling reuses the application binary rather than separate scripts.
8. **Configuration is env-driven** (`godotenv` + `caarlos0/env`); `DB_HOST/PORT/USER/PASSWORD/NAME` and `JWT_SECRET` are required, everything else has sane defaults (see `EnvConfig` in `server/main.go` and `.env.example` files).
9. **Timestamps are UTC** (`time.Time` / RFC3339); tools like dedup trust the server clock.
10. **Quality gates enforced by git hooks** — Conventional Commits, branch prefixes, no direct commits to `main`, ≥90% statement coverage on both backend and frontend.


## 6. Project description

Product Catalog is a compact full-stack web application (and boilerplate for future projects): a Vue 3 SPA talking to a Fiber REST API (`/api/v1`) backed by Postgres.

- **Catalog browsing (public)** — list products, product detail by id/name, list categories, category autocomplete via `/categories/match`.
- **Accounts** — register, login (cookie session), `/auth/me` profile, logout clearing cookies.
- **Management (authenticated)** — create/update/delete products, create categories, link/unlink categories to products, and a personal "my products" view scoped to the logged-in user.
- **Orders & maintenance** — an orders schema exists with a dedicated CLI for duplicate-order cleanup (section 4.b); order creation itself is not exposed by the API yet.
- **API documentation** — Swagger/OpenAPI generated into `server/docs/` (`make swagger`) and served by the API.

The stack: Fiber v3 + goqu + sql-migrate on the backend; Vue 3 + Pinia + vue-router + Tailwind on the frontend; Vitest and Go testing with real-database strategies on both sides (section 2.a.iv).


## 7. Project structure

### a. Backend (`server/`)

```
server/
├── commands/cmd/       # Cobra CLI: serve.go (server), migrate.go + seed.go,
│                       #   dedup_orders_remove.go; root.go wires shared config/logger
├── docs/               # Generated Swagger/OpenAPI assets (make swagger) — do not hand-edit
├── migrations/         # sql-migrate files with -- +migrate Up/Down sections
├── src/
│   ├── controllers/    # HTTP handlers: parse request → service call → JSON response
│   ├── middleware/     # AuthMiddleware.go — RequireAuth JWT guard
│   ├── models/         # DB-facing structs (one per table)
│   ├── repositories/   # goqu query builders — the only layer that touches SQL
│   ├── routes/         # v1.go registers /api/v1 groups and applies auth guards
│   ├── services/       # business logic + transactions; includes dedup_service.go
│   └── structs/        # request/response DTOs shared across layers
├── testdb/             # test helpers: in-memory SQLite w/ mirrored schema + fixtures
├── utils/              # db.go (pool), executor.go (tx wrapper), jwt.go, logger.go,
│                       #   normalize.go, uuid.go, validator.go (+ tests)
├── main.go             # loads .env → parses EnvConfig → cmd.Execute(cfg, logger)
├── makefile            # mfile/mup/mdown/mseed/dev/swagger/dedup/test/coverage/lint/format
├── .air.toml           # air watcher config used by `make dev`
├── .env.example        # template for required envs
└── .golangci.yml       # linter configuration
```

Request flow: `routes/v1.go` → `middleware.RequireAuth` (if guarded) → `controllers` → `services` (transactions via `utils.ResolveExecutor`) → `repositories` (goqu) → Postgres.

### b. Frontend (`app/`)

```
app/
├── src/
│   ├── components/
│   │   ├── Category/    # category listing/form components
│   │   ├── Product/     # ProductForm.vue, ModifyProduct.vue, CategoryLinker.vue, Index.vue
│   │   ├── layout/      # Header.vue, Footer.vue, ErrorBanner.vue
│   │   └── table/       # reusable table pieces
│   ├── layouts/Base.vue # app shell wrapping every view
│   ├── network/request.js # fetch wrapper ({ ok, data | error }) + endpoint wrappers
│   ├── router/index.js  # route table + auth guard (?redirect= on protected routes)
│   ├── stores/auth.js   # Pinia session store (login/register/logout/fetchMe)
│   ├── stores/errors.js # shared error state feeding ErrorBanner
│   └── views/           # pages: Home, Product, MyProducts, Category,
│                        #   ModifyProduct, Login, Register
├── vite.config.js       # dev server + @tailwindcss/vite plugin
├── vitest.config.js     # unit test setup (+ jsdom)
├── eslint.config.js     # ESLint config incl. Vue + oxlint interop
├── .oxlintrc.json / .prettierrc.json / tailwind.config.js
└── package.json         # scripts: dev/build/preview/test:unit/coverage/lint/format
```

Each folder with testable units carries its own `__tests__/` directory (components, views, stores, router, network).