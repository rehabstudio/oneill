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
	"github.com/rehabstudio/oneill/nginxclient"
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

	dockerClient, err := dockerclient.NewDockerClient(config.DockerApiEndpoint, config.RegistryCredentials, config.NginxDisabled)
	exitOnError(err, "Unable to initialise docker client")

	definitionLoader, err := containerdefs.GetLoader(config.DefinitionsURI)
	exitOnError(err, "Unable to load container definitions")

	definitions, err := containerdefs.LoadContainerDefinitions(definitionLoader)
	exitOnError(err, "Unable to load container definitions")

	runningContainers, err := containerdefs.ProcessContainerDefinitions(definitions, dockerClient)
	exitOnError(err, "Unable to process container definitions")

	// if nginx is disabled globally we just skip the configuration and reload steps entirely
	if !config.NginxDisabled {
		err = nginxclient.ConfigureAndReload(config, runningContainers)
		exitOnError(err, "Unable to configure and reload nginx")
	}

	err = dockerClient.RemoveOldContainers(runningContainers)
	exitOnError(err, "Unable to stop and remove old containers")

}
