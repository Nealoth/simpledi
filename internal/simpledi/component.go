package simpledi

var _ IComponent = &Component{}

type IComponent interface {
	PreInit()
	PostInit()
	OnDestroy()
}

type Component struct {
}

func (c *Component) PreInit() {
}

func (c *Component) PostInit() {
}

func (c *Component) OnDestroy() {
}
