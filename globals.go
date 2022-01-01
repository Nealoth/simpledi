package simpledi

import (
	"sync"
)

var containerInitMutex = sync.Once{}
var globalContainer IDiContainer

func init() {
	containerInitMutex.Do(func() {
		globalContainer = &defaultDiContainer{}
		globalContainer.Init()
	})
}

func CreateContainer() IDiContainer {
	return &defaultDiContainer{}
}

func Register(cmp IComponent) int {
	globalContainer.RegisterComponent(cmp)
	return 0
}

func Start() {
	globalContainer.Start()
}

func Destroy() {
	globalContainer.Destroy()
}
