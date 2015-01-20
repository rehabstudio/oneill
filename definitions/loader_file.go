package definitions

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/rehabstudio/oneill/logger"
)

type LoaderFile struct {
	path string
}

func (l *LoaderFile) ValidateURI() error {

	// check if path exists
	src, err := os.Stat(l.path)
	if err != nil {
		return err
	}

	// check if path is actually a file
	if src.IsDir() {
		return fmt.Errorf("%s is a directory", l.path)
	}

	return nil
}

// LoadContainerDefinitions reads a local yaml or json file for container definitions,
// loads them into memory and unmarshalls them into ContainerDefinition structs.
func (l *LoaderFile) LoadContainerDefinitions() ([]*ContainerDefinition, error) {
	logger.L.Debug("Loading container definitions: single file")

	var cd []*ContainerDefinition

	// read file from disk
	data, err := ioutil.ReadFile(l.path)
	if err != nil {
		return cd, err
	}

	// unmarshall yaml data to struct
	err = yaml.Unmarshal(data, &cd)
	if err != nil {
		return cd, err
	}

	return cd, nil
}
