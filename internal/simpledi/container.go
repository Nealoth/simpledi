package simpledi

import (
	"fmt"
	"github.com/Nealoth/simpledi/pkg/simpledi/reflections"
	"sort"
)

var _ IDiContainer = &DefaultDiContainer{}

type IDiContainer interface {
	Init()
	RegisterComponent(cmp IComponent)
	Start()
	GetComponentByName(name string) (IComponent, bool)
	GetComponent(component IComponent) (IComponent, bool)
	Destroy()
}

type DefaultDiContainer struct {
	initialized bool
	definitions definitionsContainer
	components  componentsContainer
}

func (d *DefaultDiContainer) GetComponent(component IComponent) (IComponent, bool) {
	return d.GetComponentByName(reflections.GetTypeFullName(component))
}

func (d *DefaultDiContainer) GetComponentByName(name string) (IComponent, bool) {
	component, found := d.components[name]

	if found {
		return component, true
	}

	return nil, false
}

func (d *DefaultDiContainer) Start() {

	errDef := "cannot start DI container"

	if !d.initialized {
		panic(fmt.Sprintf("%s: container has not been initialized", errDef))
	}

	// --- VERIFICATION STAGE. VERIFY THAT ALL OF DEPS ARE EXIST AND THERE ARE NO CIRCULAR INJECTIONS
	verificationErr := d.definitions.
		verify(
			verifyDefinitionsDependencies,
			verifyCircularDependencyInjections,
		)

	if verificationErr != nil {
		panic(fmt.Sprintf("%s: %s", errDef, verificationErr))
	}

	// --- PREPARING TO INIT LOOP
	definitionsList := make([]*componentDefinition, 0, len(d.definitions))

	for _, definition := range d.definitions {
		definitionsList = append(definitionsList, definition)
	}

	sort.Slice(definitionsList, func(i, j int) bool {
		return len(definitionsList[i].dependencies) < len(definitionsList[j].dependencies)
	})

	// --- INIT LOOP
	componentsToInitLeft := len(definitionsList)

	for componentsToInitLeft > 0 {
		componentsInitialized := 0
		for _, definition := range definitionsList {
			if !definition.initialized {

				// --- CHECK IF COMPONENT READY TO INIT

				componentReadyToInit := true

				// check deps for init ready
				for _, depName := range definition.dependencies {
					dep, depFound := d.definitions[depName]

					// if error which dep is not found
					if !depFound {
						componentReadyToInit = false
						panic(fmt.Sprintf("%s: something went wrong, dep '%s' of component '%s' has not found",
							errDef,
							depName,
							definition.fullName,
						))
					}

					// if dep not yet initialized
					if !dep.initialized {
						componentReadyToInit = false
						continue
					}
				}

				if !componentReadyToInit {
					continue
				}

				// --- INIT

				// pre init stage
				definition.rawComponent.PreInit()

				// inject stage

				for injectorName, injectorFunc := range definition.injectors {

					fieldsValues := reflections.
						GetTypeFieldsValuesByTagValue(definition.rawComponent, injectTagName, injectorName)

					if err := injectorFunc(definition, fieldsValues, d.components); err != nil {
						panic(fmt.Sprintf(
							"%s: injection error occured. component: '%s', injector: '%s', err: %s",
							errDef,
							definition.fullName,
							injectorName,
							err,
						))
					}
				}

				// post init
				definition.rawComponent.PostInit()

				componentsInitialized++
				componentsToInitLeft--
				definition.initialized = true
				d.components[definition.fullName] = definition.rawComponent
			}
		}

		if componentsInitialized == 0 {
			// Log initialized components and not initialized components
			panic(fmt.Sprintf("%s: components initialization infinite loop",
				errDef))
		}
	}

	// --- POST INIT STAGE

	// Log total initialized/not initialized

	for _, def := range d.definitions {
		if !def.initialized {
			panic(fmt.Sprintf("%s: something went wrong: component %s has not been initialized by init loop",
				errDef, def.fullName))
		}
	}

	d.purifyDefinitions()
	d.afterContainerStart()
}

func (d *DefaultDiContainer) afterContainerStart() {

	// TODO log after container start
	for _, comp := range d.components {
		comp.AfterContainerStart()
	}

}

func (d *DefaultDiContainer) purifyDefinitions() {

	for key, value := range d.definitions {
		value.rawComponent = nil
		delete(d.definitions, key)
	}

	d.definitions = nil
}

func (d *DefaultDiContainer) Init() {
	d.definitions = make(definitionsContainer, 0)
	d.components = make(componentsContainer, 0)
	d.initialized = true
}

func (d *DefaultDiContainer) RegisterComponent(cmp IComponent) {

	errDef := "cannot register component"

	if !d.initialized {
		panic(fmt.Sprintf("%s: container has not been initialized", errDef))
	}

	componentName := reflections.GetTypeFullName(cmp)

	if !reflections.IsPointer(cmp) {
		panic(fmt.Sprintf("%s: component '%s' should be registered as pointer", errDef, componentName))
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
			panic(fmt.Sprintf("%s: field '%s' of component '%s' has unknown injection type '%s'",
				errDef,
				field.Name+" "+field.Type.String(),
				componentName,
				injectionType))
		}

		injectorsMap[injectionType] = injectorFunction
		componentDependencies = append(componentDependencies, fieldTypeName)
	}

	_, componentAlreadyExist := d.definitions[componentName]

	if componentAlreadyExist {
		panic(fmt.Sprintf("%s: component '%s' already exist", errDef, componentName))
	}

	d.definitions[componentName] = &componentDefinition{
		fullName:     componentName,
		rawComponent: cmp,
		dependencies: componentDependencies,
		injectors:    injectorsMap,
	}
}

func (d *DefaultDiContainer) Destroy() {
	for _, component := range d.components {
		component.OnDestroy()
	}
}
