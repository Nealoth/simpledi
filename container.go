package simpledi

type IDiContainer interface {
	Init()
	RegisterComponent(cmp IComponent)
	Start()
	Destroy()
}
