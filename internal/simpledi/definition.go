package simpledi

type componentDefinition struct {
	fullName     string
	initialized  bool
	rawComponent IComponent
	dependencies []string
	injectors    map[string]injectionFunc
}
