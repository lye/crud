package crud

import (
	"fmt"
	"reflect"
	"database/sql"
)

func Scan(rs *sql.Rows, args ...interface{}) error {
	prefix := ""

	writeBackMap := make(map[string]interface{})

	for _, arg := range args {
		val := reflect.ValueOf(arg)
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

	cols, er := rs.Columns()
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

	if er := rs.Scan(writeBack...) ; er != nil {
		fmt.Printf("Error encountered, columns: %#v\n", cols)
		return er
	}

	return nil
}
