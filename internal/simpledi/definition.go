package simpledi

type componentDefinition struct {
	fullName     string
	initialized  bool
	rawComponent IComponent
	dependencies []string
	injectors    map[string]injectionFunc
}

type definitionsContainer map[string]*componentDefinition

func (d definitionsContainer) verify(verificators ...definitionsVerificatorFunc) error {
	for _, verificator := range verificators {
		if err := verificator(d); err != nil {
			return err
		}
	}
	return nil
}
