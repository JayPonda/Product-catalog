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