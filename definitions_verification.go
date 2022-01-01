package simpledi

import (
	"errors"
	"fmt"
)

type definitionsVerificatorFunc func(d definitionsContainer) error

func verifyDefinitionsDependencies(d definitionsContainer) error {

	errDef := "container verification error"

	// Verify that all dependencies are exist
	for cName, cDef := range d {
		for _, dep := range cDef.dependencies {
			if _, depFound := d[dep]; !depFound {
				return errors.New(fmt.Sprintf("%s. component '%s' cannot be initialized, dependepcy '%s' has not found",
					errDef,
					cName,
					dep,
				))
			}
		}
	}

	return nil
}

func verifyCircularDependencyInjections(d definitionsContainer) error {

	injectedByMap := make(map[string]map[string]bool, 0)

	// Constructing a map which will contain map of component and deps which will inject this component
	for _, def := range d {

		if _, found := injectedByMap[def.fullName]; !found {
			injectedByMap[def.fullName] = make(map[string]bool, 0)
		}

		for _, def2 := range d {
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
		for reverseDepName := range injectedBy {
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
						return errors.New(fmt.Sprintf("circular dependency detected in component '%s'", definitionName))
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

					return errors.New(fmt.Sprintf("circular dependency detected: %s", formattedFullPath))
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

	return nil
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
