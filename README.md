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
2. Tools information
    - a. Backend
        - i. migration
        - ii. orm
        - iii. uuid
        - iv. testing
    - b. Frontend
        - i. material UI
        - ii. testing
    - c. General
        - i. hooks
3. production setup
    - a. config envs
    - b. docker setup
4. scope description
    - a. form simulator
    - b. duplicate order aggrigator (Cli)
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

## Testing & coverage

Run everything from the repo root via the root Makefile:

```
$ root > make test            # backend + frontend suites
$ root > make coverage        # both coverage reports
```

### a. Backend (`server/`)

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

### b. Frontend (`app/`)

| Command | What it does |
|---|---|
| `pnpm test:unit` | Vitest in watch mode |
| `pnpm test:unit -- --run` | single pass (CI style) |
| `pnpm coverage` | V8 coverage, text + HTML (`coverage/index.html`) |

Layers covered:

1. **network** (`src/network/request.js`) — URL/method/body/credentials for every endpoint wrapper, `{ ok, data | error }` result shape on success, HTTP error statuses, network failures, and error-message extraction for auth calls
2. **stores** (`src/stores/auth.ts`) — login/register/logout/fetchMe state transitions including failure paths leaving state untouched
3. **components** — `Header.vue`: logged-in vs logged-out UI, logout clearing session + navigating, mobile menu toggle
4. **router guard** — protected-route redirects (`/my-products`, `/products/add`, `/categories/add`, `/products/:id/edit`) carry a `redirect` query; public routes stay open; session restored exactly once