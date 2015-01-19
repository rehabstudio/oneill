package definitions

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"

	"github.com/rehabstudio/oneill/logger"
)

type LoaderDirectoryPerDefinition struct {
	rootDirectory string
}

func (l *LoaderDirectoryPerDefinition) ValidateURI() error {

	// check if rootDirectory exists
	src, err := os.Stat(l.rootDirectory)
	if err != nil {
		return err
	}

	// check if rootDirectory is actually a directory
	if !src.IsDir() {
		return fmt.Errorf("%s is not a directory", l.rootDirectory)
	}

	return nil
}

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

	return &cd, nil
}

// LoadContainerDefinitions scans a local directory (might have been passed from the command line)
// for container definitions, reads them into memory and unmarshalls them into ContainerDefinition
// structs.
func (l *LoaderDirectoryPerDefinition) LoadContainerDefinitions() ([]*ContainerDefinition, error) {
	logger.L.Debug("Loading container definitions: one directory per definition")

	var cds []*ContainerDefinition

	// scan the configured directory, erroring if we don't have permission, it doesn't exist, etc.
	dirContents, err := ioutil.ReadDir(l.rootDirectory)
	if err != nil {
		return cds, err
	}

	// load all definitions contained in the configured definitions directory
	for _, f := range dirContents {
		if f.IsDir() {
			cdPath := path.Join(l.rootDirectory, f.Name(), "container.yaml")
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

	return cds, nil
}
