package utils

import (
	"database/sql"
	"testing"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "modernc.org/sqlite"
)

type fakeExec struct{}

func (fakeExec) From(...interface{}) *goqu.SelectDataset { return nil }
func (fakeExec) Insert(interface{}) *goqu.InsertDataset  { return nil }
func (fakeExec) Update(interface{}) *goqu.UpdateDataset  { return nil }
func (fakeExec) Delete(interface{}) *goqu.DeleteDataset  { return nil }

func TestResolveExecutor(t *testing.T) {
	db, err := sql.Open("sqlite", "file:exec_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gdb := goqu.New("sqlite3", db)

	if got := ResolveExecutor(gdb, nil); got != Executor(gdb) {
		t.Error("expected db when no execs provided")
	}

	fe := fakeExec{}
	if got := ResolveExecutor(gdb, []Executor{fe}); got != fe {
		t.Error("expected provided executor to be used")
	}

	if got := ResolveExecutor(gdb, []Executor{nil}); got != Executor(gdb) {
		t.Error("expected db when provided exec is nil")
	}
}
