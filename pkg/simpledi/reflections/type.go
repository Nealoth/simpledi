package reflections

import (
	"reflect"
)

func GetTypeFullName(any interface{}) string {
	t := reflect.TypeOf(any)
	return t.String()
}

func IsPointer(any interface{}) bool {
	v := reflect.ValueOf(any)
	return v.Kind() == reflect.Ptr
}

func FieldIsPointer(field reflect.StructField) bool {
	return field.Type.Kind() == reflect.Ptr
}
