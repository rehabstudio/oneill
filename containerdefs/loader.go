package containerdefs

import (
	"fmt"
)

type DefinitionLoader interface {
	LoadContainerDefinitions() ([]*ContainerDefinition, error)
	ValidateURI() error
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

	// validate all container definitions individually, dropping any that
	// don't pass validation
	var definitionsValidated []*ContainerDefinition
	for _, definition := range definitions {
		if definition.Validate() {
			definitionsValidated = append(definitionsValidated, definition)
		}
	}

	// validate container definitions as a group, if this doesn't pass then we
	// bail out since it's impossible to know what the user meant to do.
	for _, definition := range definitionsValidated {
		if !definitionIsUnique(definition, definitionsValidated) {
			return []*ContainerDefinition{}, fmt.Errorf("Container definitions clash (name or ports): %s", definition.ContainerName)
		}
	}

	return definitionsValidated, nil
}

func definitionIsUnique(cd *ContainerDefinition, cds []*ContainerDefinition) bool {

	// check for clashing container names
	var containerNameCount int
	for _, ocd := range cds {
		if ocd.ContainerName == cd.ContainerName {
			containerNameCount = containerNameCount + 1
		}
	}
	// return false if any containers with same name found
	if containerNameCount > 1 {
		return false
	}

	// check for clashing port numbers
	var portCount int
	for port, _ := range cd.PortMapping {
		port = 0
		for _, ocd := range cds {
			for oport, _ := range ocd.PortMapping {
				if oport == port {
					portCount = portCount + 1
				}
			}
		}
		// return false if any ports with same number found
		if portCount > 1 {
			return false
		}
	}

	return true
}
