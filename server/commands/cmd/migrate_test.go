package cmd

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JayPonda/Product-catalog/server/testdb"
	"github.com/doug-martin/goqu/v9"
)

// ---- migrationVersion (pure) ----

func TestMigrationVersion(t *testing.T) {
	v, ok := migrationVersion("20260816125333_create_category_table.sql")
	if !ok || v != 20260816125333 {
		t.Errorf("got %d,%v want 20260816125333,true", v, ok)
	}
	if v, ok := migrationVersion("short.sql"); ok {
		t.Errorf("short id should fail, got %d", v)
	}
	if v, ok := migrationVersion("abcdefgh1234567_not_a_number.sql"); ok {
		t.Errorf("non-numeric prefix should fail, got %d", v)
	}
}

// ---- migrationFileIDs (filesystem) ----

func TestMigrationFileIDs_OrdersByTimestampAndFilters(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"20260102030405_beta.sql":  "", // newer timestamp
		"20250101010101_alpha.sql": "", // older timestamp, lexically LATER
		"notes.txt":                "", // not a migration
		"123.sql":                  "", // too short to hold a timestamp
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir.sql"), 0755); err != nil { // dir with .sql suffix
		t.Fatal(err)
	}

	ids, err := migrationFileIDs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected only the two valid migrations, got %v", ids)
	}
	if ids[0] != "20250101010101_alpha.sql" || ids[1] != "20260102030405_beta.sql" {
		t.Errorf("timestamp order violated: %v", ids)
	}
}

func TestMigrationFileIDs_InvalidTimestampErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "20261332030405_bad_month.sql"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := migrationFileIDs(dir); err == nil {
		t.Error("invalid timestamp in filename must error")
	}
}

func TestMigrationFileIDs_MissingDirErrors(t *testing.T) {
	if _, err := migrationFileIDs(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Error("missing migrations dir must error")
	}
}

// ---- isDuplicateKey (pure) ----

func TestIsDuplicateKey(t *testing.T) {
	cases := []struct {
		msg  string
		want bool
	}{
		{`pq: duplicate key value violates unique constraint "gorp_migrations_pkey"`, true},
		{"Duplicate entry '20260101' for key 'PRIMARY'", true},
		{"UNIQUE constraint failed: gorp_migrations.id", true},
		{"connection refused", false},
	}
	for _, tc := range cases {
		if got := isDuplicateKey(errors.New(tc.msg)); got != tc.want {
			t.Errorf("isDuplicateKey(%q) = %v, want %v", tc.msg, got, tc.want)
		}
	}
}

// ---- previousLatestVersion / recordedMaxVersion (real SQLite) ----

func newRawSQLite(t *testing.T) (*sql.DB, *goqu.Database) {
	t.Helper()
	goquDB := testdb.OpenSQLite(t)
	raw, ok := goquDB.Db.(*sql.DB)
	if !ok {
		t.Fatalf("underlying goqu Db is %T, want *sql.DB", goquDB.Db)
	}
	return raw, goquDB
}

func TestPreviousLatestVersion_IgnoresDirtyAndEmpty(t *testing.T) {
	sqlDB, goquDB := newRawSQLite(t)

	if _, err := sqlDB.Exec(`CREATE TABLE schema_migrations (version bigint, dirty boolean)`); err != nil {
		t.Fatal(err)
	}
	stmt := `INSERT INTO schema_migrations (version, dirty) VALUES (?, ?)`
	for _, row := range [][2]interface{}{
		{int64(20260816125333), true},  // dirty → ignored
		{int64(20260817000000), false}, // highest clean
		{int64(20240101010101), false},
	} {
		if _, err := sqlDB.Exec(stmt, row[0], row[1]); err != nil {
			t.Fatal(err)
		}
	}

	got, err := previousLatestVersion(sqlDB)
	if err != nil {
		t.Fatal(err)
	}
	if got != 20260817000000 {
		t.Errorf("latest clean version = %d, want 20260817000000", got)
	}

	if _, err := goquDB.Delete("schema_migrations").Executor().Exec(); err != nil {
		t.Fatal(err)
	}
	if got, err = previousLatestVersion(sqlDB); err != nil || got != 0 {
		t.Errorf("empty table should give 0,nil, got %d,%v", got, err)
	}
}

func TestRecordedMaxVersion_ParsesIdsSkipsBogus(t *testing.T) {
	sqlDB, goquDB := newRawSQLite(t)

	if _, err := sqlDB.Exec(`CREATE TABLE gorp_migrations (id text, applied_at timestamp)`); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{
		"20250101010101_old.sql",
		"20260816125333_create_category_table.sql",
		"bogus_id_without_timestamp.sql", // unparsable → skipped
	} {
		if _, err := sqlDB.Exec(`INSERT INTO gorp_migrations (id, applied_at) VALUES (?, ?)`, id, time.Now()); err != nil {
			t.Fatal(err)
		}
	}

	got, err := recordedMaxVersion(sqlDB)
	if err != nil {
		t.Fatal(err)
	}
	if got != 20260816125333 {
		t.Errorf("max recorded version = %d, want 20260816125333", got)
	}

	if _, err := goquDB.Delete("gorp_migrations").Executor().Exec(); err != nil {
		t.Fatal(err)
	}
	if got, err = recordedMaxVersion(sqlDB); err != nil || got != 0 {
		t.Errorf("empty table should give 0,nil, got %d,%v", got, err)
	}
}
