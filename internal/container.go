package internal

import (
	"fmt"
	"github.com/Nealoth/simpledi"
	"github.com/Nealoth/simpledi/internal/reflections"
	"sort"
)

var _ simpledi.IDiContainer = &DefaultDiContainer{}

type DefaultDiContainer struct {
	initialized bool
	definitions definitionsContainer
	components  componentsContainer
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

					fieldsValues := reflections.GetTypeFieldsValuesByTagValue(definition.rawComponent, injectTagName, injectorName)

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
				d.components.AddComponent(definition.fullName, definition.rawComponent)
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
		comp.OnContainerReady()
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

func (d *DefaultDiContainer) RegisterComponent(cmp simpledi.IComponent) {

	errDef := "component registration error"

	if !d.initialized {
		panic(fmt.Sprintf("%s: container has not been initialized", errDef))
	}

	if _, err := d.definitions.register(cmp); err != nil {
		panic(fmt.Sprintf("%s: %s", errDef, err))
	}
}

func (d *DefaultDiContainer) Destroy() {
	for _, component := range d.components {
		component.OnDestroy()
	}
}
