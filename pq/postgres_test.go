package crud

import (
	"database/sql"
	_ "github.com/bmizerany/pq"
	"os"
	"testing"
	"time"
)

type Foo struct {
	Id   int64     `crud:"foo_id"`
	Num  int64     `crud:"foo_num"`
	Str  string    `crud:"foo_str"`
	Time time.Time `crud:"foo_time"`
}

type OptionalFoo struct {
	Int8    *int8    `crud:"o_int8"`
	Int16   *int16   `crud:"o_int16"`
	Int32   *int32   `crud:"o_int32"`
	Int64   *int64   `crud:"o_int64"`
	Float32 *float32 `crud:"o_float32"`
	Float64 *float64 `crud:"o_float64"`
	Bool    *bool    `crud:"o_bool"`
	String  *string  `crud:"o_string"`
}

type TimeFoo struct {
	Int     time.Time  `crud:"time_int,unix"`
	IntPtr  *time.Time `crud:"time_int_ptr,unix"`
	Time    time.Time  `crud:"time_val"`
	TimePtr *time.Time `crud:"time_val_ptr"`
}

func newFoo() Foo {
	return Foo{
		Num:  42,
		Str:  "PANIC",
		Time: time.Unix(1338, 0).UTC(),
	}
}

func openTestConn() (*sql.DB, error) {
	sslmode := os.Getenv("PGSSLMODE")

	if sslmode == "" {
		os.Setenv("PGSSLMODE", "disable")
	}

	return sql.Open("postgres", "")
}

func createDb() (*sql.DB, error) {
	db, er := openTestConn()
	if er != nil {
		return nil, er
	}

	_, er = db.Exec(`
		CREATE TABLE foo
			( foo_id BIGSERIAL PRIMARY KEY
			, foo_num INTEGER NOT NULL
			, foo_str VARCHAR(34) NOT NULL
			, foo_time TIMESTAMP NOT NULL
			);
	`)

	if er != nil {
		db.Close()
		return nil, er
	}

	_, er = db.Exec(`
		CREATE TABLE ofoo
			( o_int8 INTEGER
			, o_int16 INTEGER
			, o_int32 INTEGER
			, o_int64 INTEGER
			, o_float32 REAL
			, o_float64 DOUBLE PRECISION
			, o_bool BOOL
			, o_string VARCHAR(255)
			);
	`)

	if er != nil {
		db.Close()
		return nil, er
	}

	_, er = db.Exec(`
		CREATE TABLE tfoo
			( time_int INTEGER NOT NULL
			, time_int_ptr INTEGER
			, time_val TIMESTAMP NOT NULL
			, time_val_ptr TIMESTAMP
			)
	`)

	if er != nil {
		db.Close()
		return nil, er
	}

	return db, nil
}

func destroyDb(db *sql.DB) {
	_, _ = db.Exec(`DROP TABLE foo CASCADE`)
	_, _ = db.Exec(`DROP TABLE ofoo CASCADE`)
	_, _ = db.Exec(`DROP TABLE tfoo CASCADE`)
}

func TestSingleFoo(t *testing.T) {
	db, er := createDb()
	if er != nil {
		t.Fatal(er)
	}
	defer db.Close()
	defer destroyDb(db)

	f := newFoo()

	if er := Update(db, "foo", "foo_id", f); er == nil {
		t.Errorf("Expected Update to error on zero-id field")
	}

	if er := Update(db, "foo", "does_not_exist", f); er == nil {
		t.Errorf("Expected Update to error on non-existant ID field")
	}

	if er := Update(db, "foo", "foo_id", ""); er == nil {
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

	if er := Scan(rows, &f2); er != nil {
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

func TestModifyFoo(t *testing.T) {
	db, er := createDb()
	if er != nil {
		t.Fatal(er)
	}
	defer db.Close()
	defer destroyDb(db)

	f := newFoo()

	if f.Id, er = Insert(db, "foo", "foo_id", f); er != nil {
		t.Fatal(er)
	}

	f.Num = 3
	f.Str = "hello"

	if er := Update(db, "foo", "foo_id", f); er != nil {
		t.Fatal(er)
	}

	rows, er := db.Query("SELECT * FROM foo")
	if er != nil {
		t.Fatal(er)
	}

	foos := []Foo{}

	if er := ScanAll(rows, &foos); er != nil {
		t.Fatal(er)
	}

	if len(foos) != 1 {
		t.Fatalf("Got wrong number of foos: %d (expected %d)", len(foos), 1)
	}

	if foos[0].Str != "hello" {
		t.Errorf("Str mismatch: expected '%s', got '%s'", "hello", foos[0].Str)
	}

	if foos[0].Num != 3 {
		t.Errorf("Num mismatch: expected %d, got %d", 3, foos[0].Num)
	}
}

func TestModifyFooByPtr(t *testing.T) {
	db, er := createDb()
	if er != nil {
		t.Fatal(er)
	}
	defer db.Close()
	defer destroyDb(db)

	f := newFoo()

	if f.Id, er = Insert(db, "foo", "foo_id", &f); er != nil {
		t.Fatal(er)
	}

	f.Num = 3
	f.Str = "hello"

	if er := Update(db, "foo", "foo_id", &f); er != nil {
		t.Fatal(er)
	}

	rows, er := db.Query("SELECT * FROM foo")
	if er != nil {
		t.Fatal(er)
	}

	foos := []Foo{}

	if er := ScanAll(rows, &foos); er != nil {
		t.Fatal(er)
	}

	if len(foos) != 1 {
		t.Fatalf("Got wrong number of foos: %d (expected %d)", len(foos), 1)
	}

	if foos[0].Str != "hello" {
		t.Errorf("Str mismatch: expected '%s', got '%s'", "hello", foos[0].Str)
	}

	if foos[0].Num != 3 {
		t.Errorf("Num mismatch: expected %d, got %d", 3, foos[0].Num)
	}
}

func TestMultipleFoo(t *testing.T) {
	db, er := createDb()
	if er != nil {
		t.Fatal(er)
	}
	defer db.Close()
	defer destroyDb(db)

	f1 := Foo{
		Num: 3,
	}

	f2 := Foo{
		Num: 12,
	}

	if f1.Id, er = Insert(db, "foo", "foo_id", f1); er != nil {
		t.Fatal(er)
	}

	foos := []Foo{}
	rows, er := db.Query("SELECT * FROM foo")
	if er != nil {
		t.Fatal(er)
	}

	if er := ScanAll(rows, &foos); er != nil {
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

	if f2.Id, er = Insert(db, "foo", "foo_id", f2); er != nil {
		t.Fatal(er)
	}

	foos = []Foo{}
	rows, er = db.Query("SELECT * FROM foo")
	if er != nil {
		t.Fatal(er)
	}

	if er := ScanAll(rows, &foos); er != nil {
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

func TestOptionalFoo(t *testing.T) {
	db, er := createDb()
	if er != nil {
		t.Fatal(er)
	}
	defer db.Close()
	defer destroyDb(db)

	var Int8 int8 = 8
	var Int16 int16 = 16
	var Int32 int32 = 32
	var Int64 int64 = 64
	var Float32 float32 = 0.32
	var Float64 float64 = 0.64
	var Bool bool = true
	var String string = "string"

	f1 := OptionalFoo{
		Int8:    &Int8,
		Int16:   &Int16,
		Int32:   &Int32,
		Int64:   &Int64,
		Float32: &Float32,
		Float64: &Float64,
		Bool:    &Bool,
		String:  &String,
	}

	if _, er := Insert(db, "ofoo", "", f1); er != nil {
		t.Fatal(er)
	}

	rows, er := db.Query("SELECT * FROM ofoo")
	if er != nil {
		t.Fatal(er)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatalf("No rows returned? wtf")
	}

	f2 := OptionalFoo{}

	if er := Scan(rows, &f2); er != nil {
		t.Fatal(er)
	}

	if f2.Int8 == nil {
		t.Errorf("Int8 - nil")

	} else if *f2.Int8 != Int8 {
		t.Errorf("Int8 - mismatch")
	}

	if f2.Int16 == nil {
		t.Errorf("Int16 - nil")

	} else if *f2.Int16 != Int16 {
		t.Errorf("Int16 - mismatch")
	}

	if f2.Int32 == nil {
		t.Errorf("Int32 - nil")

	} else if *f2.Int32 != Int32 {
		t.Errorf("Int32 - mismatch")
	}

	if f2.Int64 == nil {
		t.Errorf("Int64 - nil")

	} else if *f2.Int64 != Int64 {
		t.Errorf("Int64 - mismatch")
	}

	if f2.Float32 == nil {
		t.Errorf("Float32 - nil")

	} else if *f2.Float32 != Float32 {
		t.Errorf("Float32 - mismatch")
	}

	if f2.Float64 == nil {
		t.Errorf("Float64 - nil")

	} else if *f2.Float64 != Float64 {
		t.Errorf("Float64 - mismatch")
	}

	if f2.Bool == nil {
		t.Errorf("Bool - nil")

	} else if *f2.Bool != Bool {
		t.Errorf("Bool - mismatch")
	}

	if f2.String == nil {
		t.Errorf("String - nil")

	} else if *f2.String != String {
		t.Errorf("String - mismatch")
	}
}

func TestNullOptionalFoo(t *testing.T) {
	db, er := createDb()
	if er != nil {
		t.Fatal(er)
	}
	defer db.Close()
	defer destroyDb(db)

	f1 := OptionalFoo{}

	if _, er := Insert(db, "ofoo", "", f1); er != nil {
		t.Fatal(er)
	}

	rows, er := db.Query("SELECT * FROM ofoo")
	if er != nil {
		t.Fatal(er)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("No rows returned?")
	}

	f2 := OptionalFoo{}

	if er := Scan(rows, &f2); er != nil {
		t.Fatal(er)
	}

	if f2.Int8 != nil {
		t.Errorf("Int8 - not nil")
	}

	if f2.Int16 != nil {
		t.Errorf("Int16 - not nil")
	}

	if f2.Int32 != nil {
		t.Errorf("Int32 - not nil")
	}

	if f2.Int64 != nil {
		t.Errorf("Int64 - not nil")
	}

	if f2.Float32 != nil {
		t.Errorf("Float32 - not nil")
	}

	if f2.Float64 != nil {
		t.Errorf("Float64 - not nil")
	}

	if f2.Bool != nil {
		t.Errorf("Bool - not nil")
	}

	if f2.String != nil {
		t.Errorf("String - not nil")
	}
}

func TestTimeMarshalling(t *testing.T) {
	db, er := createDb()
	if er != nil {
		t.Fatal(er)
	}
	defer destroyDb(db)

	/* Ugh, convert to Unix for granularity reasons; strip off TZ data
	 * because SQLite doesn't understand them and they affect both .Equal 
	 * and .Unix */
	now := time.Unix(time.Now().Unix(), 0).UTC()

	foo1 := TimeFoo{
		Int:     now,
		IntPtr:  &now,
		Time:    now,
		TimePtr: &now,
	}

	foo2 := TimeFoo{}

	testEqual := func() {
		if foo1.Int.Unix() != foo2.Int.Unix() {
			t.Errorf("mismatch - Int, e: %d, a: %d", foo1.Int.Unix(), foo2.Int.Unix())
		}

		if foo1.IntPtr.Unix() != foo2.IntPtr.Unix() {
			t.Errorf("mismatch - IntPtr, e: %d, a: %d", foo1.IntPtr.Unix(), foo2.IntPtr.Unix())
		}

		if !foo1.Time.Equal(foo2.Time) {
			t.Errorf("mismatch - Time\ne: %s\na: %s", foo1.Time.String(), foo2.Time.String())
		}

		if !foo1.TimePtr.Equal(*foo2.TimePtr) {
			t.Errorf("mismatch - TimePtr\ne: %s\na: %s", foo1.TimePtr.String(), foo2.TimePtr.String())
		}
	}

	if _, er := Insert(db, "tfoo", "", foo1); er != nil {
		t.Fatal(er)
	}

	rows, er := db.Query("SELECT time_int, time_int_ptr, time_val, time_val_ptr FROM tfoo")
	if er != nil {
		t.Fatal(er)
	}

	if !rows.Next() {
		rows.Close()
		t.Errorf("Insert inserted no rows?")

	} else {
		var tmpInt int64
		var tmpIntPtr sql.NullInt64

		if er := rows.Scan(&tmpInt, &tmpIntPtr, &foo2.Time, &foo2.TimePtr); er != nil {
			rows.Close()
			t.Error(er)

		} else {
			tmp := time.Unix(tmpIntPtr.Int64, 0)
			foo2.IntPtr = &tmp
			foo2.Int = time.Unix(tmpInt, 0)

			testEqual()
		}

		rows.Close()
	}

	foo2 = TimeFoo{}

	rows, er = db.Query("SELECT * FROM tfoo")
	if er != nil {
		t.Fatal(er)
	}

	if !rows.Next() {
		rows.Close()
		t.Errorf("Rows are gone?")

	} else {
		if er := Scan(rows, &foo2); er != nil {
			rows.Close()
			t.Error(er)

		} else {
			rows.Close()
			testEqual()
		}
	}
}
