package simpledi

var _ IComponent = &Component{}

type IComponent interface {
	PreInit()
	PostInit()
}

type Component struct {
	componentName string
}

func (c *Component) PreInit() {
}

func (c *Component) PostInit() {
}
