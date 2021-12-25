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
		f.Set(reflect.ValueOf(components[f.Type().String()]))
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
		fieldsComparisonMap[strings.ReplaceAll(argType.String(), "*", "")] = true
	}

	for _, actualField := range fields {
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
			fmt.Sprintf("arguments mismatch. '%s' func  args '%s' have not been declared as injectable",
				funcInjectName,
				strings.Join(argsNames, ", "),
			))
	}

	injectionArgs := make([]reflect.Value, 0, len(funcArgsTypes))
	injectionArgs = append(injectionArgs, reflect.ValueOf(definition.rawComponent))

	for _, argType := range funcArgsTypes {
		component, found := components[argType.String()]

		if !found {
			return errors.New(fmt.Sprintf("something went wrong. component '%s' has not found", argType.String()))
		}

		injectionArgs = append(injectionArgs, reflect.ValueOf(component))
	}

	injectFunc.Func.Call(injectionArgs)

	return nil
}
