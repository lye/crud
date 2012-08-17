package crud

import (
	"fmt"
	"reflect"
	"database/sql"
)

/*
Scan uses tag metadata and column names to extract values from a sql.Rows into one or more objects.

Scan inspects all of the passed arguments and creates a mapping from SQL column
name to the fields of the passed objects by inspecting the struct tag metadata.
It then constructs an appropriate call to rows.Scan, passing in pointers as the
mapping dictates.

Any string passed in the arguments list is considered a "prefix" for the SQL
names of each field in the preceding object.

If two objects have fields that map to the same column name, only the first is
assigned properly. If two columns have the same SQL name, the same interface is
passed for both fields (and which gets bound is undefined). If there is a SQL 
column which does not map to a Go field (or vice versa), it is ignored silently.
*/
func Scan(rows *sql.Rows, args ...interface{}) error {
	prefix := ""

	writeBackMap := make(map[string]interface{})

	for _, arg := range args {
		val := indirectV(reflect.ValueOf(arg))
		ty := val.Type()

		if ty.Kind() == reflect.String {
			prefix = arg.(string)
			continue
		}

		fieldMap, er := sqlToGoFields(ty)
		if er != nil {
			return er
		}

		for sqlName, goName := range fieldMap {
			sqlName = prefix + sqlName
			writeBackMap[sqlName] = val.FieldByName(goName).Addr().Interface()
		}

		prefix = ""
	}

	cols, er := rows.Columns()
	if er != nil {
		return er
	}

	writeBack := make([]interface{}, len(cols))

	for i, col := range cols {
		if target, ok := writeBackMap[col] ; ok {
			writeBack[i] = target

		} else {
			writeBack[i] = new(interface{})
		}
	}

	if er := rows.Scan(writeBack...) ; er != nil {
		fmt.Printf("Error encountered, columns: %#v\n", cols)
		return er
	}

	return nil
}

/*
ScanAll accepts a pointer to a slice of a type and fills it with repeated calls to Scan.

ScanAll only works if you're trying to extract a single object from each row
of the query results. Additionally, it closes the passed sql.Rows object. ScanAll
effectively replaces this code

  // old code
  defer rows.Close()
  objs := []Object{}
  for rows.Next() {
	  var obj Object
	  Scan(rows, &obj)
	  objs = append(objs, obj)
  }

With simply

  // new code
  objs := []Object{}
  ScanAll(rows, &objs)
*/
func ScanAll(rows *sql.Rows, slicePtr interface{}) error {
	defer rows.Close()

	sliceVal := reflect.ValueOf(slicePtr).Elem()

	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("Argument to crud.ScanAll is not a slice")
	}

	elemType := sliceVal.Type().Elem()

	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("Argument to crud.ScanAll must be a slice of structs")
	}

	for rows.Next() {
		newVal := reflect.New(elemType)

		if er := Scan(rows, newVal.Interface()) ; er != nil {
			return er
		}

		sliceVal.Set(reflect.Append(sliceVal, newVal.Elem()))
	}

	return nil
}

