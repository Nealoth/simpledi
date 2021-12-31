package internal

import (
	"fmt"
	"github.com/Nealoth/simpledi"
)

type componentsContainer map[string]simpledi.IComponent

func (c componentsContainer) GetRequired(name string) simpledi.IComponent {
	if component, found := c[name]; found {
		return component
	} else {
		panic(fmt.Sprintf("component '%s' not found", name))
	}
}

func (c componentsContainer) AddComponent(componentName string, component simpledi.IComponent) {
	if _, found := c[componentName]; !found {
		c[componentName] = component
	} else {
		panic(fmt.Sprintf("cannot add component '%s', component already exist", componentName))
	}
}
