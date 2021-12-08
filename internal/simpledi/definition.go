package simpledi

type DefinitionRegistryType int

type ComponentDefinition struct {
	fullName     string
	initialized  bool
	rawComponent IComponent
	dependencies []string
}
