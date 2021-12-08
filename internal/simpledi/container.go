package simpledi

import (
	"fmt"
	"github.com/Nealoth/simpledi/pkg/simpledi/reflections"
	"reflect"
	"sort"
)

var globalContainer DiContainer

type definitionsMap map[string]*ComponentDefinition
type componentsMap map[string]IComponent

var _ DiContainer = &DefaultDiContainer{}

type DiContainer interface {
	Init()
	RegisterComponent(cmp IComponent)
	Start()
	GetComponentByName(name string) (IComponent, bool)
	GetComponent(component IComponent) (IComponent, bool)
}

type DefaultDiContainer struct {
	initialized bool
	definitions definitionsMap
	components  componentsMap
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
	d.verifyDefinitionsDependencies()
	d.verifyCircularDependencyInjections()

	// --- PREPARING TO INIT LOOP
	definitionsList := make([]*ComponentDefinition, 0)

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
				v := reflect.ValueOf(definition.rawComponent).Elem()

				fields := make([]reflect.Value, 0)

				for i := 0; i < v.NumField(); i++ {
					fields = append(fields, v.Field(i))
				}

				for _, depName := range definition.dependencies {
					for _, field := range fields {
						if depName == field.Type().String() {
							// TODO check exportable and non exportable. For not it will work only for exportable field
							field.Set(reflect.ValueOf(d.components[depName]))
						}
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

}

func (d *DefaultDiContainer) Init() {
	d.definitions = make(definitionsMap, 0)
	d.components = make(componentsMap, 0)
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
	componentDependencies := make([]string, 0)

	for _, field := range injectableFields {
		if reflections.FieldIsPointer(field) {
			componentDependencies = append(componentDependencies, field.Type.String())
		} else {
			panic(fmt.Sprintf("%s: field '%s' of component '%s' should be a pointer",
				errDef,
				field.Name+" "+field.Type.String(),
				componentName,
			))
		}
	}

	_, componentAlreadyExist := d.definitions[componentName]

	if componentAlreadyExist {
		panic(fmt.Sprintf("%s: component '%s' already exist", errDef, componentName))
	}

	d.definitions[componentName] = &ComponentDefinition{
		fullName:     componentName,
		rawComponent: cmp,
		dependencies: componentDependencies,
	}
}

func (d *DefaultDiContainer) verifyDefinitionsDependencies() {

	errDef := "container verification error"

	// Verify that all dependencies are exist
	for cName, cDef := range d.definitions {
		for _, dep := range cDef.dependencies {
			if _, depFound := d.definitions[dep]; !depFound {
				panic(fmt.Sprintf("%s: component '%s' cannot be initialized, dependepcy '%s' not found",
					errDef,
					cName,
					dep,
				))
			}
		}
	}
}

func (d *DefaultDiContainer) verifyCircularDependencyInjections() {

	injectedByMap := make(map[string]map[string]bool, 0)

	// Constructing a map which will contain map of component and deps which will inject this component
	for _, def := range d.definitions {

		if _, found := injectedByMap[def.fullName]; !found {
			injectedByMap[def.fullName] = make(map[string]bool, 0)
		}

		for _, def2 := range d.definitions {
			for _, dep := range def2.dependencies {
				if dep == def.fullName {
					injectedByMap[def.fullName][def2.fullName] = true
				}
			}
		}
	}

	// Circular injections search loop start
	for definitionName, injectedBy := range injectedByMap {

		// Buffer for processing deps
		injectedByBuf := make([]string, 0)

		// Add first iteration to buffer
		for reverseDepName, _ := range injectedBy {
			injectedByBuf = append(injectedByBuf, reverseDepName)
		}

		// Iterate when elements are exist in buffer
		for len(injectedByBuf) > 0 {

			newBuf := make([]string, 0)

			for _, injectedByInner := range injectedByBuf {

				// If some dep (or dep of dep) of initial component contain this initial component as dependency
				if injectedByInner == definitionName {

					// Path for print where actual loop is
					path := append(make([]string, 0), injectedByInner)

					// Running recursive search of loop point, like DFS
					fullPath, found := findLoopPath(0, injectedByInner, path, injectedByMap)

					if !found {
						// TODO make error message more detailed
						panic("something went wrong")
					}

					// Pretty format found path
					formattedFullPath := ""

					for i := len(fullPath) - 1; i >= 0; i-- {
						separator := " -> "

						if i == 0 {
							separator = ""
						}

						formattedFullPath += fullPath[i] + separator
					}

					panic("LOOP DETECTED! " + formattedFullPath)
				}

				// Fill new buf by current component deps
				for outerInjectedByDepName := range injectedByMap[injectedByInner] {
					newBuf = append(newBuf, outerInjectedByDepName)
				}
			}

			// Updating a buf if it has not found on current iteration
			injectedByBuf = newBuf
		}

	}
}

func findLoopPath(currentIteration int, loopDetectedAt string, path []string, injectedBy map[string]map[string]bool) ([]string, bool) {

	errDef := "injection loop detector error"

	if currentIteration > 25 {
		panic(fmt.Sprintf("%s: max injection loop find hops (%d hops) exceeded", errDef, 25))
	}

	for k := range injectedBy[path[len(path)-1]] {

		if k == loopDetectedAt {
			return append(path, loopDetectedAt), true
		}

		newPath := make([]string, 0)
		newPath = append(newPath, path...)
		newPath = append(newPath, k)

		foundPath, found := findLoopPath(currentIteration+1, loopDetectedAt, newPath, injectedBy)

		if found {
			return foundPath, true
		}
	}

	return nil, false
}
