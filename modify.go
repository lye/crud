package crud

import (
	"fmt"
	"reflect"
	"strings"
)

/* 
Update syncs a tagged object with an existing record in the database.
 
The metadata contained in the crud tags don't include the table name or
the name of the SQL primary ID, so they have to be passed in manually.
If the object passed in as arg does not have a primary key set (or the
value is 0), an error is returned.
*/
func Update(db DbIsh, table, sqlIdFieldName string, arg interface{}) error {
	val := reflect.ValueOf(arg)
	ty := val.Type()

	fieldMap, er := sqlToGoFields(ty)
	if er != nil {
		return er
	}

	sqlFields := make([]string, len(fieldMap))[:0]
	newValues := make([]interface{}, len(fieldMap))[:0]
	var id int64 = 0

	for sqlName, goName := range fieldMap {
		if sqlName == sqlIdFieldName {
			id = val.FieldByName(goName).Int()

		} else {
			sqlFields = append(sqlFields, sqlName)
			newValues = append(newValues, val.FieldByName(goName).Interface())
		}
	}

	if id == 0 {
		return fmt.Errorf("%s is 0 or not set, cannot update", sqlIdFieldName)
	}

	newValues = append(newValues, id)

	q := fmt.Sprintf("UPDATE %s SET %s WHERE %s = $1", table, strings.Join(sqlFields, ", "), sqlIdFieldName)
	_, er = db.Exec(q, newValues...)
	return er
}

/* 
Insert creates a new record in the datastore.
*/
func Insert(db DbIsh, table, sqlIdFieldName string, arg interface{}) (int64, error) {
	val := reflect.ValueOf(arg)
	ty := val.Type()

	fieldMap, er := sqlToGoFields(ty)
	if er != nil {
		return 0, er
	}

	sqlFields := make([]string, len(fieldMap))[:0]
	newValues := make([]interface{}, len(fieldMap))[:0]
	placeholders := make([]string, len(fieldMap))[:0]

	for sqlName, goName := range fieldMap {
		if sqlName == sqlIdFieldName {
			continue
		}

		sqlFields = append(sqlFields, sqlName)
		newValues = append(newValues, val.FieldByName(goName).Interface())
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(placeholders) + 1))
	}

	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(sqlFields, ", "), strings.Join(placeholders, ", "))

	res, er := db.Exec(q, newValues...)
	if er != nil {
		return 0, er
	}

	return res.LastInsertId()
}
