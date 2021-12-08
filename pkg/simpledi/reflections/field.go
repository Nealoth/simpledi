package reflections

import "reflect"

func GetTypeFields(any interface{}) []reflect.StructField {
	t := reflect.TypeOf(any)

	e := t.Elem()

	fields := make([]reflect.StructField, 0)

	for i := 0; i < e.NumField(); i++ {
		fields = append(fields, e.Field(i))
	}

	return fields
}

func GetTypeFieldsByTag(any interface{}, tag string) []reflect.StructField {
	t := reflect.TypeOf(any)

	e := t.Elem()

	fields := make([]reflect.StructField, 0)

	for i := 0; i < e.NumField(); i++ {
		field := e.Field(i)

		if _, found := field.Tag.Lookup(tag); found {
			fields = append(fields, field)
		}

	}

	return fields
}
