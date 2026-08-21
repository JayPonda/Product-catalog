package utils

import (
	"github.com/doug-martin/goqu/v9"
)

type Executor interface {
	From(cols ...interface{}) *goqu.SelectDataset
	Insert(table interface{}) *goqu.InsertDataset
	Update(table interface{}) *goqu.UpdateDataset
	Delete(table interface{}) *goqu.DeleteDataset
}

func ResolveExecutor(db *goqu.Database, execs []Executor) Executor {
	if len(execs) > 0 && execs[0] != nil {
		return execs[0]
	}
	return db
}
