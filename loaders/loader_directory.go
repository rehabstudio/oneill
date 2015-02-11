package loaders

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/rehabstudio/oneill/containerdefs"
)

type LoaderDirectory struct {
	rootDirectory string
}

func (l *LoaderDirectory) ValidateURI() error {
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

// LoadContainerDefinitions scans a local directory (might have been passed from the command line)
// for container definitions, reads them into memory and unmarshalls them into ContainerDefinition
// structs. This scan is not recursive and will not search subdirectories for definitions.
func (l *LoaderDirectory) LoadContainerDefinitions() ([]*containerdefs.ContainerDefinition, error) {
	logrus.WithFields(logrus.Fields{
		"source": "directory",
		"path":   l.rootDirectory,
	}).Debug("Loading container definitions")

	var cds []*containerdefs.ContainerDefinition

	// scan the configured directory, erroring if we don't have permission, it doesn't exist, etc.
	dirContents, err := ioutil.ReadDir(l.rootDirectory)
	if err != nil {
		return cds, err
	}

	// load all definitions contained in the configured definitions directory
	for _, f := range dirContents {
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if ext == ".yaml" || ext == ".json" {
			cdPath := path.Join(l.rootDirectory, f.Name())
			var cd *containerdefs.ContainerDefinition
			cd, err := loadSingleContainerDefinition(cdPath)
			// if we aren't able to load the definition for some reason we just move on to the next
			// folder, it's not fatal, oneill will just act as if it doesn't exist
			if err != nil {
				continue
			}
			logrus.WithFields(logrus.Fields{"path": cdPath}).Debug("Found container definition")
			cds = append(cds, cd)
		}
	}

	return cds, nil
}

// loadSingleContainerDefinition loads a single container definition from disk and unmarshalls
// it into a ContainerDefinition struct
func loadSingleContainerDefinition(path string) (*containerdefs.ContainerDefinition, error) {

	cd := containerdefs.ContainerDefinition{}

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
