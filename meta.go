package crud

import (
	"fmt"
	"reflect"
)

func sqlToGoFields(ty reflect.Type) (map[string]string, error) {
	ty = indirect(ty)

	if ty.Kind() != reflect.Struct {
		return nil, fmt.Errorf("sqlToGoFields: type %s is not a struct", ty.Name())
	}

	fieldMap := make(map[string]string)

	for i := 0 ; i < ty.NumField() ; i += 1 {
		field := ty.Field(i)

		tag := field.Tag.Get("crud")

		if tag != "" {
			fieldMap[tag] = field.Name
		}
	}

	return fieldMap, nil
}

func indirect(ty reflect.Type) reflect.Type {
	for ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}

	return ty
}
