package crud

import (
	"fmt"
	"reflect"
	"strings"
)

type fieldMeta struct {
	GoName  string
	SqlName string
	Unix    bool
}

/* 
sqlToGoFields extracts the data contained within the "crud" struct tags.

The "crud" struct tags contain the name of the SQL column which stores the
tagged field. sqlToGoFields returns a mapping from SQL column name to Go
field name.
*/
func sqlToGoFields(ty reflect.Type) (map[string]fieldMeta, error) {
	ty = indirectT(ty)

	if ty.Kind() != reflect.Struct {
		return nil, fmt.Errorf("sqlToGoFields: type %s is not a struct", ty.Name())
	}

	fieldMap := make(map[string]fieldMeta)

	for i := 0; i < ty.NumField(); i += 1 {
		field := ty.Field(i)

		tag := field.Tag.Get("crud")

		if tag != "" {
			tagPieces := strings.Split(tag, ",")

			meta := fieldMeta{
				SqlName: tagPieces[0],
				GoName:  field.Name,
			}

			for idx := 1; idx < len(tagPieces); idx += 1 {
				if tagPieces[idx] == "unix" {
					meta.Unix = true
				}
			}

			fieldMap[meta.SqlName] = meta
		}
	}

	return fieldMap, nil
}

/* indirectT returns the passed type, recursively indirected. */
func indirectT(ty reflect.Type) reflect.Type {
	for ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}

	return ty
}

/* indirectV returns the passed type, recursively indirected. */
func indirectV(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return val
}
