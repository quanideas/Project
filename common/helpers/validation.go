package helpers

import (
	"errors"
	"reflect"
)

func ValidateFieldGetAllQuery(fieldName string, object interface{}) (bool, error) {
	obj := reflect.ValueOf(object)

	// Iterate through each field
	for i := 0; i < obj.Type().NumField(); i++ {
		objField := obj.Type().Field(i)

		// Json name of field found (json name matches db field name)
		if jsonTag := objField.Tag.Get("json"); jsonTag != "" && jsonTag != "-" && jsonTag == fieldName {
			if objField.Type.Kind() == reflect.Bool {
				return true, nil
			} else {
				return false, nil
			}
		}
	}

	return false, errors.New("field not found")
}
