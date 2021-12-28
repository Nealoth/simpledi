package simpledi

import (
	"errors"
	"fmt"
	"github.com/Nealoth/simpledi/pkg/simpledi/reflections"
	"reflect"
	"strings"
)

const funcInjectName = "Inject"

type injectionFunc func(definition *componentDefinition, fields []reflect.Value, components componentsContainer) error

var injectValues = map[string]injectionFunc{
	"field": fieldInject,
	"func":  funcInject,
}

func fieldInject(_ *componentDefinition, fields []reflect.Value, components componentsContainer) error {

	for _, f := range fields {
		if !f.CanSet() {
			return errors.New(fmt.Sprintf("cannot set value to field with type: '%s'", f.Type().String()))
		}

		if f.Kind() == reflect.Ptr {
			f.Set(reflect.ValueOf(components[f.Type().String()]))
		} else {
			f.Set(reflect.ValueOf(components["*"+f.Type().String()]).Elem())
		}
	}

	return nil
}

func funcInject(definition *componentDefinition, fields []reflect.Value, components componentsContainer) error {

	defType := reflect.TypeOf(definition.rawComponent)

	injectFunc, injectFuncFound := defType.MethodByName(funcInjectName)

	if !injectFuncFound {
		return errors.New(fmt.Sprintf("injection func '%s' has not found", funcInjectName))
	}

	funcArgsTypes := reflections.GetFuncArgsTypes(injectFunc)[1:]

	// Arguments count validation
	if len(funcArgsTypes) == 0 {
		return errors.New(fmt.Sprintf("injection func without args is redundant use 'PreInit' or 'PostInit instead'"))
	}

	// Arguments count comparison with expected validation
	if len(fields) != len(funcArgsTypes) {
		return errors.New(fmt.Sprintf("%s function arguments mismatch. expected %d arguments but got %d",
			funcInjectName,
			len(fields),
			len(funcArgsTypes),
		))
	}

	// Arguments types comparison with expected validation
	fieldsComparisonMap := make(map[string]bool)

	for _, argType := range funcArgsTypes {
		// Replace pointers to compare struct and pointer types
		fieldsComparisonMap[strings.ReplaceAll(argType.String(), "*", "")] = true
	}

	for _, actualField := range fields {
		// Replace pointers to compare struct and pointer types
		delete(fieldsComparisonMap, strings.ReplaceAll(actualField.Type().String(), "*", ""))
	}

	if len(fieldsComparisonMap) > 0 {
		argsNames := make([]string, 0)

		for argName := range fieldsComparisonMap {
			if argName != "" {
				argsNames = append(argsNames, argName)
			}
		}

		return errors.New(
			fmt.Sprintf("arguments mismatch. '%s' func  arg(s) '%s' have not been declared as injectable field(s) for '%s'",
				funcInjectName,
				strings.Join(argsNames, ", "),
				definition.fullName,
			))
	}

	injectionArgs := make([]reflect.Value, 0, len(funcArgsTypes))
	injectionArgs = append(injectionArgs, reflect.ValueOf(definition.rawComponent))

	for _, argType := range funcArgsTypes {
		if argType.Kind() == reflect.Ptr {
			component, found := components[argType.String()]

			if !found {
				return errors.New(fmt.Sprintf("something went wrong. component '%s' has not found", argType.String()))
			}

			injectionArgs = append(injectionArgs, reflect.ValueOf(component))
		} else {
			component, found := components["*"+argType.String()]

			if !found {
				return errors.New(fmt.Sprintf("something went wrong. component '%s' has not found", argType.String()))
			}

			injectionArgs = append(injectionArgs, reflect.ValueOf(component).Elem())
		}
	}

	injectFunc.Func.Call(injectionArgs)

	return nil
}
