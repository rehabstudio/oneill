package definitions

// GetLoader parses a given URI and returns an appropriate loader. For now
// this always returns our default (and only) loader, but could be easily
// expanded to load container definitions from a remote location, or from a
// single file.
func GetLoader(uri string) (DefinitionLoader, error) {
	return &LoaderDirectoryPerDefinition{rootDirectory: uri}, nil
}

// LoadContainerDefinitions scans a local directory (might have been passed from the command line)
// for container definitions, reads them into memory and unmarshalls them into ContainerDefinition
// structs.
func LoadContainerDefinitions(loader DefinitionLoader) ([]*ContainerDefinition, error) {

	// validate the uri that's been passed to the definition, this might be ensuring that a given
	// directory exists or that a url returns a 200 status code.
	if err := loader.ValidateURI(); err != nil {
		return []*ContainerDefinition{}, err
	}

	// load container definitions. By default this is from disk, but could be from a remote
	// location if a loader for that source exists.
	definitions, err := loader.LoadContainerDefinitions()
	if err != nil {
		return definitions, err
	}

	// load default values for any fields in the container definition not set by the loader
	for i := range definitions {
		loadContainerDefaults(definitions[i])
	}

	// validate all container definitions, dropping any that don't pass validation
	var definitionsValidated []*ContainerDefinition
	for _, definition := range definitions {
		if validateDefinition(definition, definitions) {
			definitionsValidated = append(definitionsValidated, definition)
		}
	}

	return definitionsValidated, nil
}
