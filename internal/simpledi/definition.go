package simpledi

import (
	"errors"
	"fmt"
	"github.com/Nealoth/simpledi/pkg/simpledi/reflections"
)

type componentDefinition struct {
	fullName     string
	initialized  bool
	rawComponent IComponent
	dependencies []string
	injectors    map[string]injectionFunc
}

type definitionsContainer map[string]*componentDefinition

func (d definitionsContainer) register(cmp IComponent) (*componentDefinition, error) {
	componentName := reflections.GetTypeFullName(cmp)

	if !reflections.IsPointer(cmp) {
		return nil, errors.New(fmt.Sprintf("component '%s' should be registered as pointer", componentName))
	}

	_, componentAlreadyExist := d[componentName]

	if componentAlreadyExist {
		return nil, errors.New(fmt.Sprintf("component '%s' already exist", componentName))
	}

	injectableFields := reflections.GetTypeFieldsByTag(cmp, injectTagName)
	componentDependencies := make([]string, 0, len(injectableFields))
	injectorsMap := make(map[string]injectionFunc, 0)

	for _, field := range injectableFields {

		fieldTypeName := field.Type.String()

		if !reflections.FieldIsPointer(field) {
			fieldTypeName = "*" + fieldTypeName
		}

		injectionType, _ := field.Tag.Lookup(injectTagName)

		injectorFunction, exist := injectValues[injectionType]

		if !exist {
			return nil, errors.New(fmt.Sprintf("field '%s' of component '%s' has unknown injection type '%s'",
				field.Name+" "+field.Type.String(),
				componentName,
				injectionType))
		}

		injectorsMap[injectionType] = injectorFunction
		componentDependencies = append(componentDependencies, fieldTypeName)
	}

	component := &componentDefinition{
		fullName:     componentName,
		rawComponent: cmp,
		dependencies: componentDependencies,
		injectors:    injectorsMap,
	}

	d[componentName] = component

	return component, nil
}

func (d definitionsContainer) verify(verificators ...definitionsVerificatorFunc) error {
	for _, verificator := range verificators {
		// TODO add trace logging
		if err := verificator(d); err != nil {
			return err
		}
	}
	return nil
}
