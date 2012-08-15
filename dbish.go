package crud

import "database/sql"

type DbIsh interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Prepare(string, ...interface{}) (*sql.Stmt, error)
	Query(string, ...interface{}) (*sql.Rows, error)
}
