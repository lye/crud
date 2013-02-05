package crud

import "database/sql"

/* 
DbIsh provides an interface that is implemented by both sql.DB and sql.Tx. 

All crud methods accept DbIsh's to allow the end-user to use the interface both
within and without transactions.
*/
type DbIsh interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Prepare(string) (*sql.Stmt, error)
	Query(string, ...interface{}) (*sql.Rows, error)
}
