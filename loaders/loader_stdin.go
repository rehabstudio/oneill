package loaders

import (
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/rehabstudio/oneill/containerdefs"
)

type LoaderStdin struct{}

// we don't have a uri to validate for this loader
func (l *LoaderStdin) ValidateURI() error {
	return nil
}

// LoadContainerDefinitions reads yaml or json data from stdin.
func (l *LoaderStdin) LoadContainerDefinitions() ([]*containerdefs.ContainerDefinition, error) {
	logrus.WithFields(logrus.Fields{
		"source": "stdin",
		"path":   nil,
	}).Debug("Loading container definitions")

	var cd []*containerdefs.ContainerDefinition

	// read data from stdin
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return cd, err
	}

	// unmarshall yaml/json data to struct
	err = yaml.Unmarshal(data, &cd)
	if err != nil {
		return cd, err
	}

	return cd, nil
}
