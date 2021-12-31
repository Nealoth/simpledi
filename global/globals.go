package global

import (
	"github.com/Nealoth/simpledi"
	"github.com/Nealoth/simpledi/internal"
	"sync"
)

var containerInitMutex = sync.Once{}
var globalContainer simpledi.IDiContainer

func init() {
	containerInitMutex.Do(func() {
		globalContainer = &internal.DefaultDiContainer{}
		globalContainer.Init()
	})
}

func Register(cmp simpledi.IComponent) int {
	globalContainer.RegisterComponent(cmp)
	return 0
}

func Start() {
	globalContainer.Start()
}

func Destroy() {
	globalContainer.Destroy()
}
