package crud

import (
	"time"
	"testing"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Foo struct {
	Id int64 `crud:"foo_id"`
	Num int64 `crud:"foo_num"`
	Str string `crud:"foo_str"`
	Time time.Time `crud:"foo_time"`
}

func newFoo() Foo {
	return Foo{
		Num: 42,
		Str: "PANIC",
		Time: time.Unix(1338, 0).UTC(),
	}
}

func createDb() (*sql.DB, error) {
	db, er := sql.Open("sqlite3", ":memory:")
	if er != nil {
		return nil, er
	}

	_, er = db.Exec(`
		CREATE TABLE foo
			( foo_id INTEGER PRIMARY KEY AUTOINCREMENT
			, foo_num INTEGER NOT NULL
			, foo_str VARCHAR(34) NOT NULL
			, foo_time TIMESTAMP NOT NULL
			)
	`)

	if er != nil {
		db.Close()
		return nil, er
	}

	return db, er
}

func TestSingleFoo(t *testing.T) {
	db, er := createDb()
	if er != nil {
		t.Fatal(er)
	}
	defer db.Close()

	f := newFoo()

	if er := Update(db, "foo", "foo_id", f) ; er == nil {
		t.Errorf("Expected Update to error on zero-id field")
	}

	if er := Update(db, "foo", "does_not_exist", f) ; er == nil {
		t.Errorf("Expected Update to error on non-existant ID field")
	}

	if er := Update(db, "foo", "foo_id", "") ; er == nil {
		t.Errorf("Expected Update to error on non-struct type")
	}

	f.Id, er = Insert(db, "foo", "foo_id", f)
	if er != nil {
		t.Fatal(er)
	}

	if f.Id == 0 {
		t.Fatalf("Expected Insert to return non-0 id (got %d)", f.Id)
	}

	var f2 Foo

	rows, er := db.Query("SELECT * FROM foo")
	if er != nil {
		t.Fatal(er)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatalf("No rows appear to have been inserted")
	}

	if er := Scan(rows, &f2) ; er != nil {
		t.Fatal(er)
	}

	if f.Id != f2.Id {
		t.Errorf("Scan mismatch, ID: %d != %d", f.Id, f2.Id)
	}

	if f.Num != f2.Num {
		t.Errorf("Scan mismatch, Num: %d != %d", f.Num, f2.Num)
	}

	if f.Str != f2.Str {
		t.Errorf("Scan mismatch, Str: %d != %d", f.Str, f2.Str)
	}

	if !f.Time.Equal(f2.Time) {
		t.Errorf("Scan mismatch, Time: %v != %v", f.Time, f2.Time)
	}
}

func TestMultipleFoo(t *testing.T) {
	db, er := createDb()
	if er != nil {
		t.Fatal(er)
	}
	defer db.Close()

	f1 := Foo{
		Num: 3,
	}

	f2 := Foo{
		Num: 12,
	}

	if f1.Id, er = Insert(db, "foo", "foo_id", f1) ; er != nil {
		t.Fatal(er)
	}

	foos := []Foo{}
	rows, er := db.Query("SELECT * FROM foo")
	if er != nil {
		t.Fatal(er)
	}

	if er := ScanAll(rows, &foos) ; er != nil {
		t.Fatal(er)
	}

	if len(foos) != 1 {
		t.Fatalf("Incorrect number of foos returned from first query, got %#v\n", foos)
	}

	if foos[0].Id != f1.Id {
		t.Errorf("ScanAll mismatch: Id: %d != %d", f1.Id, foos[0].Id)
	}

	if foos[0].Num != 3 {
		t.Errorf("ScanAll mismatch: Num: %d != %d", f1.Num, foos[0].Num)
	}

	if f2.Id, er = Insert(db, "foo", "foo_id", f2) ; er != nil {
		t.Fatal(er)
	}

	foos = []Foo{}
	rows, er = db.Query("SELECT * FROM foo")
	if er != nil {
		t.Fatal(er)
	}

	if er := ScanAll(rows, &foos) ; er != nil {
		t.Fatal(er)
	}

	if len(foos) != 2 {
		t.Fatalf("Incorrect number of foos returned from second query, got %#v\n", foos)
	}

	for _, foo := range foos {
		if foo.Id == f1.Id {
			if foo.Num != f1.Num {
				t.Errorf("ScanAll mismatch: Num: %d != %d", f1.Num, foo.Num)
			}

			f1.Id = 0

		} else if foo.Id == f2.Id {
			if foo.Num != f2.Num {
				t.Errorf("ScanAll mismatch: Num: %d != %d", f2.Num, foo.Num)
			}

			f2.Id = 0

		} else {
			t.Errorf("Got unknown foo from ScanAll: %#v\n", foo)
		}
	}
}
