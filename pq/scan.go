package crud

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"
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

	intRemap := make(map[reflect.Value]*sql.NullInt64)
	floatRemap := make(map[reflect.Value]*sql.NullFloat64)
	boolRemap := make(map[reflect.Value]*sql.NullBool)
	stringRemap := make(map[reflect.Value]*sql.NullString)
	unixTimeRemap := make(map[reflect.Value]*sql.NullInt64)

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

		for sqlName, meta := range fieldMap {
			goName := meta.GoName
			sqlName = prefix + sqlName

			field := val.FieldByName(goName)
			fieldType := field.Type()

			if meta.Unix {
				nullInt := new(sql.NullInt64)
				writeBackMap[sqlName] = nullInt
				unixTimeRemap[field] = nullInt
				continue

			} else if fieldType.Kind() == reflect.Ptr {
				fieldElemKind := fieldType.Elem().Kind()

				switch fieldElemKind {
				case reflect.Int8:
					fallthrough
				case reflect.Int16:
					fallthrough
				case reflect.Int32:
					fallthrough
				case reflect.Int64:
					nullInt := new(sql.NullInt64)
					writeBackMap[sqlName] = nullInt
					intRemap[field] = nullInt
					continue

				case reflect.Float32:
					fallthrough
				case reflect.Float64:
					nullFloat := new(sql.NullFloat64)
					writeBackMap[sqlName] = nullFloat
					floatRemap[field] = nullFloat
					continue

				case reflect.Bool:
					nullBool := new(sql.NullBool)
					writeBackMap[sqlName] = nullBool
					boolRemap[field] = nullBool
					continue

				case reflect.String:
					nullString := new(sql.NullString)
					writeBackMap[sqlName] = nullString
					stringRemap[field] = nullString
					continue
				}
			}

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
		if target, ok := writeBackMap[col]; ok {
			writeBack[i] = target

		} else {
			writeBack[i] = new(interface{})
		}
	}

	if er := rows.Scan(writeBack...); er != nil {
		fmt.Printf("Error encountered, columns: %#v\n", cols)
		return er
	}

	for field, nullInt := range intRemap {
		if nullInt.Valid {
			switch field.Type().Elem().Kind() {
			case reflect.Int8:
				tmp := int8(nullInt.Int64)
				field.Set(reflect.ValueOf(&tmp))

			case reflect.Int16:
				tmp := int16(nullInt.Int64)
				field.Set(reflect.ValueOf(&tmp))

			case reflect.Int32:
				tmp := int32(nullInt.Int64)
				field.Set(reflect.ValueOf(&tmp))

			case reflect.Int64:
				field.Set(reflect.ValueOf(&nullInt.Int64))
			}
		}
	}

	for field, nullFloat := range floatRemap {
		if nullFloat.Valid {
			switch field.Type().Elem().Kind() {
			case reflect.Float32:
				tmp := float32(nullFloat.Float64)
				field.Set(reflect.ValueOf(&tmp))

			case reflect.Float64:
				field.Set(reflect.ValueOf(&nullFloat.Float64))
			}
		}
	}

	for field, nullBool := range boolRemap {
		if nullBool.Valid {
			field.Set(reflect.ValueOf(&nullBool.Bool))
		}
	}

	for field, nullString := range stringRemap {
		if nullString.Valid {
			field.Set(reflect.ValueOf(&nullString.String))
		}
	}

	for field, nullInt := range unixTimeRemap {
		if nullInt.Valid {
			t := time.Unix(nullInt.Int64, 0)

			if field.Kind() == reflect.Ptr && field.Type().Elem() == reflect.TypeOf(time.Time{}) {
				if field.IsNil() {
					newVal := &time.Time{}
					field.Set(reflect.ValueOf(newVal))
				}

				field = field.Elem()
			}

			if field.Type() != reflect.TypeOf(time.Time{}) {
				return fmt.Errorf("Cannot map a unix time to a non-time field (%T)", field.Interface())
			}

			field.Set(reflect.ValueOf(t))
		}
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

		if er := Scan(rows, newVal.Interface()); er != nil {
			return er
		}

		sliceVal.Set(reflect.Append(sliceVal, newVal.Elem()))
	}

	return nil
}
