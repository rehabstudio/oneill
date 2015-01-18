package definitions

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"

	"gopkg.in/yaml.v2"

	"github.com/rehabstudio/oneill/logger"
)

var (
	rxContainerName = regexp.MustCompile(`^/?[a-zA-Z0-9_-]+$`)
)

// loadContainerDefinition loads a single container.yaml file from disk and unmarshalls
// it into a ContainerDefinition struct
func loadContainerDefinition(path string) (*ContainerDefinition, error) {

	cd := ContainerDefinition{}

	// read file from disk
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return &cd, err
	}

	// unmarshall yaml data to struct
	err = yaml.Unmarshal(data, &cd)
	if err != nil {
		return &cd, err
	}

	return loadContainerDefaults(&cd)
}

// loadContainerDefaults fills in any blanks in the definition
func loadContainerDefaults(cd *ContainerDefinition) (*ContainerDefinition, error) {

	if cd.Tag == "" {
		cd.Tag = "latest"
	}

	return cd, nil
}

func definitionIsUnique(cd *ContainerDefinition, cds []*ContainerDefinition) bool {
	var count int
	for _, ocd := range cds {
		if ocd.Subdomain == cd.Subdomain {
			count = count + 1
		}
	}
	return count <= 1
}

func validateDefinition(cd *ContainerDefinition, cds []*ContainerDefinition) bool {
	logger.L.Debug(fmt.Sprintf("Validating container definition: %s", cd.Subdomain))

	if len(cd.Subdomain) < 3 {
		logger.L.Warning(fmt.Sprintf("%s is not long enough (must be at least 3 characters long)", cd.Subdomain))
		return false
	}
	if cd.Image == "" {
		logger.L.Warning(fmt.Sprintf("`Image` cannot be blank in container definition: %s", cd.Subdomain))
		return false
	}
	if !rxContainerName.MatchString(cd.Subdomain) {
		logger.L.Warning(fmt.Sprintf("%s is not a valid container name", cd.Subdomain))
		return false
	}
	if !definitionIsUnique(cd, cds) {
		logger.L.Warning(fmt.Sprintf("%s is not unique", cd.Subdomain))
		return false
	}

	return true
}

// LoadContainerDefinitions scans a local directory (might have been passed from the command line)
// for container definitions, reads them into memory and unmarshalls them into ContainerDefinition
// structs.
func LoadContainerDefinitions(definitionsDirectory string) ([]*ContainerDefinition, error) {
	logger.L.Debug("Loading container definitions")

	var cds []*ContainerDefinition
	var vcds []*ContainerDefinition

	// scan the configured directory, erroring if we don't have permission, it doesn't exist, etc.
	dirContents, err := ioutil.ReadDir(definitionsDirectory)
	if err != nil {
		return cds, err
	}

	// load all definitions contained in the configured definitions directory
	for _, f := range dirContents {
		if f.IsDir() {
			cdPath := path.Join(definitionsDirectory, f.Name(), "container.yaml")
			cd, err := loadContainerDefinition(cdPath)
			// if we aren't able to load the definition for some reason we just move on to the next
			// folder, it's not fatal, oneill will just act as if it doesn't exist
			if err != nil {
				continue
			}
			logger.L.Debug(fmt.Sprintf("Found container definition: %s", cdPath))
			cds = append(cds, cd)
		}
	}

	// validate all container definitions, dropping any that don't pass validation
	for _, cd := range cds {
		if validateDefinition(cd, cds) {
			vcds = append(vcds, cd)
		}
	}

	return vcds, nil
}
