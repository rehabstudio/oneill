package containerdefs

import (
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type LoaderURL struct {
	url string
}

func (l *LoaderURL) ValidateURI() error {
	return nil
}

// LoadContainerDefinitions reads a remote url that returns a list of container
// definitions in yaml or json format, loads them into memory and unmarshalls them
// into ContainerDefinition structs.
func (l *LoaderURL) LoadContainerDefinitions() ([]*ContainerDefinition, error) {
	logrus.WithFields(logrus.Fields{
		"source": "url",
		"path":   l.url,
	}).Debug("Loading container definitions")

	var cd []*ContainerDefinition

	response, err := http.Get(l.url)
	if err != nil {
		return cd, err
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return cd, err
	}

	// unmarshall yaml data to struct
	err = yaml.Unmarshal([]byte(data), &cd)
	if err != nil {
		return cd, err
	}

	return cd, nil
}
