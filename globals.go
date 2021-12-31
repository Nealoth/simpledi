package simpledi

import (
	"github.com/Nealoth/simpledi/internal"
	"sync"
)

var containerInitMutex = sync.Once{}
var globalContainer IDiContainer

func init() {
	containerInitMutex.Do(func() {
		globalContainer = &internal.DefaultDiContainer{}
		globalContainer.Init()
	})
}

func CreateContainer() IDiContainer {
	return &internal.DefaultDiContainer{}
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
