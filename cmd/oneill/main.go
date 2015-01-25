// oneill is a small command line application designed to manage a set of
// docker containers on a single host
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rehabstudio/oneill/config"
	"github.com/rehabstudio/oneill/containers"
	"github.com/rehabstudio/oneill/definitions"
	"github.com/rehabstudio/oneill/logger"
	"github.com/rehabstudio/oneill/oneill"
)

// exitOnError checks that an error is not nil. If the passed value is an
// error, it is logged and the program exits with an error code of 1
func exitOnError(err error, prefix string) {
	if err != nil {
		fmt.Printf("%s [FATAL]: %s: %s\n", time.Now().UTC().Format("2006-01-02T15:04:05.000Z"), prefix, err)
		os.Exit(1)
	}
}

func main() {

	// load configuration data
	config, err := config.LoadConfig()
	exitOnError(err, "Unable to load configuration")

	// initialise logger once configuration is complete
	logger.InitLogger(config.LogLevel)
	logger.L.Debug("Configuration successfully loaded")

	// connect to docker API and return a client instance
	dockerClient, err := containers.NewDockerClient(config.DockerApiEndpoint, config.RegistryCredentials)
	exitOnError(err, "Unable to connect to docker API")

	// load and validate container definitions
	definitionLoader, err := definitions.GetLoader(config.DefinitionsURI)
	exitOnError(err, "Unable to load container definitions")
	definitions, err := definitions.LoadContainerDefinitions(definitionLoader)
	exitOnError(err, "Unable to load container definitions")

	// initialise a new Application instance and start it
	application := oneill.NewApplication(config, dockerClient, definitions)
	exitOnError(application.RunApplication(), "Critical runtime error")

}
