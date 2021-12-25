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

func GetTypeFieldsByTagValue(any interface{}, tag, tagValue string) []reflect.StructField {
	t := reflect.TypeOf(any)

	e := t.Elem()

	fields := make([]reflect.StructField, 0)

	for i := 0; i < e.NumField(); i++ {
		field := e.Field(i)

		if foundVal, found := field.Tag.Lookup(tag); found {
			if foundVal == tagValue {
				fields = append(fields, field)
			}
		}

	}

	return fields
}

func GetTypeFieldsValuesByTagValue(any interface{}, tag, tagValue string) []reflect.Value {
	typeFields := GetTypeFieldsByTagValue(any, tag, tagValue)

	fieldsValues := GetFieldsValues(any)

	fields := make([]reflect.Value, 0, len(typeFields))

	for _, tf := range typeFields {
		for _, fv := range fieldsValues {
			if tf.Type.String() == fv.Type().String() {
				fields = append(fields, fv)
			}
		}
	}

	return fields
}
