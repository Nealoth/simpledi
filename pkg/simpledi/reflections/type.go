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

func GetFieldsValues(any interface{}) []reflect.Value {
	v := reflect.ValueOf(any).Elem()

	fields := make([]reflect.Value, 0, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		fields = append(fields, v.Field(i))
	}

	return fields
}

func GetFuncArgsTypes(method reflect.Method) []reflect.Type {

	inCnt := method.Type.NumIn()

	params := make([]reflect.Type, 0, inCnt)

	for i := 0; i < inCnt; i++ {
		params = append(params, method.Type.In(i))
	}

	return params
}
