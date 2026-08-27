package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JayPonda/Product-catalog/server/utils"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
)

// migrationsDir is the on-disk location of the SQL migration files, relative to
// the server module root (where the binary is launched from, e.g. via make).
const migrationsDir = "migrations"

// previousMigrationTable is the table used by the tool we are migrating away
// from (golang-migrate). It records successfully applied migration versions.
const previousMigrationTable = "schema_migrations"

// migrationTimestampLayout is the timestamp prefix format used in migration
// filenames, e.g. 20260816125333_create_category_table.sql.
const migrationTimestampLayout = "20060102150405"

// migrationFileIDs returns the migration IDs (filename without .sql), ordered
// by the timestamp prefix parsed as a time.Time (not lexical string order).
func migrationFileIDs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations dir: %w", err)
	}

	type scanned struct {
		id string
		ts time.Time
	}
	var migs []scanned
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		// sql-migrate uses the FULL filename (including .sql) as the migration
		// Id, so we must keep it when recording into gorp_migrations.
		id := e.Name()
		if len(id) < 14 {
			continue
		}
		ts, err := time.Parse(migrationTimestampLayout, id[:14])
		if err != nil {
			return nil, fmt.Errorf("invalid migration filename %q: %w", e.Name(), err)
		}
		migs = append(migs, scanned{id: id, ts: ts})
	}

	sort.Slice(migs, func(i, j int) bool { return migs[i].ts.Before(migs[j].ts) })

	ids := make([]string, len(migs))
	for i, m := range migs {
		ids[i] = m.id
	}
	return ids, nil
}

// migrationVersion extracts the numeric version (timestamp prefix) from a
// migration ID, e.g. 20260816125333 from "20260816125333_create_category_table".
func migrationVersion(id string) (int64, bool) {
	if len(id) < 14 {
		return 0, false
	}
	v, err := strconv.ParseInt(id[:14], 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// previousLatestVersion returns the highest successfully applied migration
// version recorded by the previous tool (golang-migrate), or 0 if the previous
// table is missing / empty (fresh database).
func previousLatestVersion(db *sql.DB) (int64, error) {
	var latest sql.NullInt64
	q := fmt.Sprintf("SELECT MAX(version) FROM %s WHERE dirty = false", previousMigrationTable)
	if err := db.QueryRow(q).Scan(&latest); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read previous migrations: %w", err)
	}
	if !latest.Valid {
		return 0, nil
	}
	return latest.Int64, nil
}

// recordedMaxVersion returns the highest migration version already recorded in
// sql-migrate's gorp_migrations table (parsed from the timestamp prefix of each
// id), or 0 if the table is empty.
func recordedMaxVersion(db *sql.DB) (int64, error) {
	rows, err := db.Query("SELECT id FROM gorp_migrations")
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read gorp_migrations: %w", err)
	}
	defer rows.Close()

	var max int64
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		if v, ok := migrationVersion(id); ok && v > max {
			max = v
		}
	}
	return max, rows.Err()
}

// migrateCmd is the parent command grouping all database migration tasks.
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands (up / down / create / seed)",
}

// migrateUpCmd applies every pending migration. Because sql-migrate wraps each
// migration in a transaction and only records the version after success, a
// failing file is rolled back and never marked applied — re-running `up` simply
// retries it. There is no "dirty" state and no `force` step required.
var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigrations(migrate.Up)
	},
}

// migrateDownCmd rolls back the most recent migration (configurable via --steps).
var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last migration (use --steps N for more)",
	RunE: func(cmd *cobra.Command, args []string) error {
		steps, _ := cmd.Flags().GetInt("steps")
		return runMigrations(migrate.Down, steps)
	},
}

// migrateSeedCmd marks the pre-existing (golang-migrate-era) migrations as
// applied. It is idempotent: already-recorded migrations are skipped, and it
// never touches migrations outside transitionMigrationIDs.
var migrateSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "One-time: register already-applied migrations (golang-migrate transition)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSeed()
	},
}

// migrateCreateCmd scaffolds a new empty migration file with Up/Down markers.
var migrateCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new migration file (e.g. migrate create add_users_table)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		timestamp := time.Now().Format("20060102150405")
		path := fmt.Sprintf("%s/%s_%s.sql", migrationsDir, timestamp, name)

		content := "-- +migrate Up\n\n-- +migrate Down\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create migration file: %w", err)
		}

		fmt.Printf("Created migration: %s\n", path)
		return nil
	},
}

// openDB opens a dedicated connection using the configured dialect + DSN.
func openDB() (*sql.DB, error) {
	db, err := sql.Open(appConfig.GetDialect(), appConfig.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database unreachable: %w", err)
	}
	return db, nil
}

// runMigrations opens a connection and delegates to sql-migrate. It never needs
// a `force`: on error it prints the reason and stops, leaving the database in
// the last known-good state.
func runMigrations(direction migrate.MigrationDirection, steps ...int) error {
	ctx := utils.NewCLIContext()
	db, err := openDB()
	if err != nil {
		appLogger.Error(ctx, "migrate.go", "runMigrations", "failed to open database", nil, err.Error())
		return err
	}
	defer db.Close()

	source := &migrate.FileMigrationSource{Dir: migrationsDir}

	var applied int

	if direction == migrate.Down && len(steps) > 0 && steps[0] > 0 {
		applied, err = migrate.ExecMax(db, appConfig.GetDialect(), source, migrate.Down, steps[0])
	} else {
		applied, err = migrate.Exec(db, appConfig.GetDialect(), source, direction)
	}
	if err != nil {
		appLogger.Error(ctx, "migrate.go", "runMigrations", "migration failed", utils.LoggerMeta{"direction": fmt.Sprintf("%v", direction)}, err.Error())
		return fmt.Errorf("migration %v failed: %w", direction, err)
	}

	appLogger.Info(ctx, "migrate.go", "runMigrations", "migrations completed", utils.LoggerMeta{"applied": applied, "direction": fmt.Sprintf("%v", direction)})
	return nil
}

// runSeed ports the migrations that were already applied by the previous tool
// (golang-migrate) into sql-migrate's gorp_migrations table. It replicates the
// applied set exactly: it takes the latest version already recorded in
// gorp_migrations and the latest version applied by the previous tool, then
// (after sorting the migration files by timestamp) records every migration in
// that range using sql-migrate's own API. On a fresh database (no previous
// table) it does nothing — use `migrate up`.
func runSeed() error {
	ctx := utils.NewCLIContext()
	db, err := openDB()
	if err != nil {
		appLogger.Error(ctx, "migrate.go", "runSeed", "failed to open database", nil, err.Error())
		return err
	}
	defer db.Close()

	dialect := appConfig.GetDialect()
	source := &migrate.FileMigrationSource{Dir: migrationsDir}

	_, dbMap, err := migrate.PlanMigration(db, dialect, source, migrate.Up, 0)
	if err != nil {
		appLogger.Error(ctx, "migrate.go", "runSeed", "failed to initialize migration table", nil, err.Error())
		return fmt.Errorf("failed to initialize migration table: %w", err)
	}

	recordedLatest, err := recordedMaxVersion(db)
	if err != nil {
		appLogger.Error(ctx, "migrate.go", "runSeed", "failed to read gorp_migrations", nil, err.Error())
		return err
	}

	previousLatest, err := previousLatestVersion(db)
	if err != nil {
		appLogger.Error(ctx, "migrate.go", "runSeed", "failed to read previous migrations", nil, err.Error())
		return err
	}
	if previousLatest == 0 {
		appLogger.Info(ctx, "migrate.go", "runSeed", "no previous migrations found, nothing to seed", nil)
		return nil
	}

	ids, err := migrationFileIDs(migrationsDir)
	if err != nil {
		appLogger.Error(ctx, "migrate.go", "runSeed", "failed to read migration files", nil, err.Error())
		return err
	}

	var seeded int
	for _, id := range ids {
		version, ok := migrationVersion(id)
		if !ok {
			continue
		}
		if version <= recordedLatest {
			continue
		}
		if version > previousLatest {
			continue
		}
		if err := dbMap.Insert(&migrate.MigrationRecord{Id: id, AppliedAt: time.Now()}); err != nil {
			if isDuplicateKey(err) {
				continue
			}
			appLogger.Error(ctx, "migrate.go", "runSeed", "failed to seed migration", utils.LoggerMeta{"id": id}, err.Error())
			return fmt.Errorf("failed to seed migration %s: %w", id, err)
		}
		seeded++
	}

	if seeded == 0 {
		appLogger.Info(ctx, "migrate.go", "runSeed", "gorp_migrations already up to date", nil)
	} else {
		appLogger.Info(ctx, "migrate.go", "runSeed", "seed completed", utils.LoggerMeta{"seeded": seeded})
	}
	return nil
}

// isDuplicateKey reports whether err is a unique/primary-key violation, which
// lets `mseed` be re-run safely (idempotent).
func isDuplicateKey(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "Duplicate entry") ||
		strings.Contains(msg, "UNIQUE")
}

func init() {
	migrateDownCmd.Flags().Int("steps", 1, "number of migrations to roll back")
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateSeedCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	rootCmd.AddCommand(migrateCmd)
}
