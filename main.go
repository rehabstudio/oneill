// oneill is a small command line application designed to manage a set of
// docker containers on a single host
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"

	"github.com/rehabstudio/oneill/config"
	"github.com/rehabstudio/oneill/containerdefs"
	"github.com/rehabstudio/oneill/dockerclient"
	"github.com/rehabstudio/oneill/loaders"
)

var (
	buildDate     string
	version       string
	gitBranch     string
	gitRevision   string
	gitRepoStatus string
)

// exitOnError checks that an error is not nil. If the passed value is an
// error, it is logged and the program exits with an error code of 1
func exitOnError(err error, prefix string) {
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal(prefix)
	}
}

// parseCliArgs parses any arguments passed to oneill on the command line
func parseCliArgs() (string, bool) {

	// parse config file location from command line flag
	configFilePath := flag.String("config", "/etc/oneill/config.yaml", "location of the oneill config file")
	showVersion := flag.Bool("v", false, "show version details and exit")
	flag.Parse()

	return *configFilePath, *showVersion
}

func main() {

	configFilePath, showVersion := parseCliArgs()
	if showVersion {
		fmt.Printf("oneill v%s\n\n", version)
		fmt.Printf("buildDate:     %s\n", buildDate)
		fmt.Printf("gitBranch:     %s\n", gitBranch)
		fmt.Printf("gitRevision:   %s\n", gitRevision)
		fmt.Printf("gitRepoStatus: %s\n", gitRepoStatus)
		os.Exit(0)
	}

	config, err := config.LoadConfig(configFilePath)
	exitOnError(err, "Unable to load configuration")

	logLevel, err := logrus.ParseLevel(config.LogLevel)
	exitOnError(err, "Unable to initialise logger")

	// configure global logger instance
	logrus.SetLevel(logLevel)
	if config.LogFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	err = dockerclient.InitDockerClient(config.DockerApiEndpoint, config.RegistryCredentials)
	exitOnError(err, "Unable to initialise docker client")

	// load container definitions
	definitionLoader, err := loaders.GetLoader(config.DefinitionsURI)
	exitOnError(err, "Unable to load container definitions")
	definitions, err := containerdefs.LoadContainerDefinitions(definitionLoader)
	exitOnError(err, "Unable to load container definitions")

	// stop redundant containers
	err = containerdefs.RemoveRedundantContainers(definitions)
	exitOnError(err, "Unable to remove redundant containers")

	// process all container definitions
	err = containerdefs.ProcessContainerDefinitions(definitions)
	exitOnError(err, "Unable to process service container definitions")
}
