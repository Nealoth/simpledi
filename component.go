package simpledi

import (
	"fmt"
)

type componentsContainer map[string]IComponent

func (c componentsContainer) GetRequired(name string) IComponent {
	if component, found := c[name]; found {
		return component
	} else {
		panic(fmt.Sprintf("component '%s' not found", name))
	}
}

func (c componentsContainer) AddComponent(componentName string, component IComponent) {
	if _, found := c[componentName]; !found {
		c[componentName] = component
	} else {
		panic(fmt.Sprintf("cannot add component '%s', component already exist", componentName))
	}
}
