package simpledi

type componentsContainer map[string]IComponent

var _ IComponent = &Component{}

type IComponent interface {
	PreInit()
	PostInit()
	OnDestroy()
	AfterContainerStart()
}

type Component struct {
}

func (c *Component) AfterContainerStart() {
}

func (c *Component) PreInit() {
}

func (c *Component) PostInit() {
}

func (c *Component) OnDestroy() {
}
