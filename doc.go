/*
Package crud provides a reflection-based wrapper around database/sql for common operations.

crud allows you to annotate struct fields with corresponding SQL field names.
Types annotated as such can then be easily inserted/updated/scanned from 
sql/database connections without incurring the usual programmer overhead.
crud does not handle schema generation; it simply reduces the amount of
boilerplate you'd need to write to interact with an existing schema.

Consider this case:

	type Foo struct {
		Id int64 `crud:"foo_id"`
		Num int64 `crud:"foo_num"`
		Str string `crud:"foo_str"`
		Time time.Time `crud:"foo_time"`
	}

And this existing schema:

	CREATE TABLE foo
		( foo_id INTEGER PRIMARY KEY
		, foo_num INTEGER NOT NULL
		, foo_str VARCHAR(24) NOT NULL
		, foo_time TIMESTAMP NOT NULL
		);

With vanilla database/sql, to extract values from an *sql.Rows you have to do a
considerable amount of hoop-jumping:

	// old code
	rows, _ := db.Query("SELECT foo_id, foo_num, foo_str, foo_time FROM foos")
	defer rows.Close()

	foos := []Foos{}
	for rows.Next() {
		var foo Foo
		rows.Scan(&foo.Id, &foo.Num, &foo.Str, &foo.Time)
		foos = append(foos, foo)
	}

With a significant number of columns to extract one-by-one (or extracting
multiple objects from, e.g., a complex query with lots of JOINs) the amount of
noisy code increases significantly. crud provides the following alternative:

	// new code
	rows, _ := db.Query("SELECT * FROM foos")
	foos := []Foos{}
	crud.ScanAll(rows, &foos)

The magic of reflection handles the rest.

Despite the wonder, crud is not without drawbacks. As with all interfaces that
use reflection internally, you lose both performance (untested as to how much)
and type safety.
*/
package crud
