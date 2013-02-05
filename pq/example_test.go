package crud

import (
	"time"
)

func ExampleInsert() {
	db, er := createDb()
	if er != nil {
		return
	}
	defer db.Close()
	defer destroyDb(db)

	/* Foo is defined elsewhere and has appropriately-tagged fields */
	f := Foo{
		Num:  42,
		Str:  "etc",
		Time: time.Now(),
	}

	/* "foo" is the SQL table name, "foo_id" is the primary key for this type */
	f.Id, er = Insert(db, "foo", "foo_id", f)

	if er != nil {
		/* Handle the error */
	}
}

func ExampleUpdate() {
	db, er := createDb()
	if er != nil {
		return
	}
	defer db.Close()
	defer destroyDb(db)

	/* f is a Foo instance obtained from somewhere and modified */
	f := Foo{
		Id:  4,
		Num: 49,
	}

	/* "foo" is the SQL table name, "foo_id" is the primary key for this type */
	if er := Update(db, "foo", "foo_id", f); er != nil {
		/* Handle the error */
	}
}

func ExampleScan() {
	db, er := createDb()
	if er != nil {
		return
	}
	defer db.Close()
	defer destroyDb(db)

	rows, er := db.Query("SELECT * FROM foo")
	if er != nil {
		/* Handle error */
	}
	defer rows.Close()

	fs := []Foo{}

	for rows.Next() {
		var f Foo

		if er := Scan(rows, &f); er != nil {
			/* Handle error */
		}

		fs = append(fs, f)
	}
}
