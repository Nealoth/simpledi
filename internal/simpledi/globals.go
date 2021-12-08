package simpledi

import (
	"sync"
)

var containerInitMutex = sync.Once{}
var _ = Init()

func Init() int {
	containerInitMutex.Do(func() {
		globalContainer = &DefaultDiContainer{}
		globalContainer.Init()
	})

	return 0
}

func Register(cmp IComponent) int {
	globalContainer.RegisterComponent(cmp)
	return 0
}

func GetComponentByName(name string) (IComponent, bool) {
	return globalContainer.GetComponentByName(name)
}

func GetComponent(component IComponent) (IComponent, bool) {
	return globalContainer.GetComponent(component)
}

func Start() {
	globalContainer.Start()
}