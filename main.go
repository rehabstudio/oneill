// oneill is a small command line application designed to manage a set of
// docker containers on a single host
package main

import (
	"github.com/Sirupsen/logrus"

	"github.com/rehabstudio/oneill/application"
	"github.com/rehabstudio/oneill/config"
	"github.com/rehabstudio/oneill/containers"
	"github.com/rehabstudio/oneill/definitions"
)

// exitOnError checks that an error is not nil. If the passed value is an
// error, it is logged and the program exits with an error code of 1
func exitOnError(err error, prefix string) {
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal(prefix)
	}
}

// initialises logrus logger at the specified logging level
func initLogger(levelStr string) {

	// parse level string and return an actual `Level` value that logrus can use
	level, err := logrus.ParseLevel(levelStr)
	exitOnError(err, "Unable to initialise logger")

	// Only log the specified severity or above.
	logrus.SetLevel(level)
}

func main() {

	// load configuration data
	config, err := config.LoadConfig()
	exitOnError(err, "Unable to load configuration")

	// initialise logger once configuration is complete
	initLogger(config.LogLevel)

	// connect to docker API and return a client instance
	dockerClient, err := containers.NewDockerClient(config.DockerApiEndpoint, config.RegistryCredentials)
	exitOnError(err, "Unable to connect to docker API")

	// load and validate container definitions
	definitionLoader, err := definitions.GetLoader(config.DefinitionsURI)
	exitOnError(err, "Unable to load container definitions")
	definitions, err := definitions.LoadContainerDefinitions(definitionLoader)
	exitOnError(err, "Unable to load container definitions")

	// initialise a new Application instance and start it
	application := application.NewApplication(config, dockerClient, definitions)
	exitOnError(application.RunApplication(), "Critical runtime error")

}
