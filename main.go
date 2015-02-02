// oneill is a small command line application designed to manage a set of
// docker containers on a single host
package main

import (
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/rehabstudio/oneill/config"
	"github.com/rehabstudio/oneill/containerdefs"
	"github.com/rehabstudio/oneill/dockerclient"
	"github.com/rehabstudio/oneill/nginxclient"
)

// exitOnError checks that an error is not nil. If the passed value is an
// error, it is logged and the program exits with an error code of 1
func exitOnError(err error, prefix string) {
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal(prefix)
	}
}

func main() {

	config, err := config.LoadConfig()
	exitOnError(err, "Unable to load configuration")

	logLevel, err := logrus.ParseLevel(config.LogLevel)
	exitOnError(err, "Unable to initialise logger")
	logrus.SetLevel(logLevel)

	dockerClient, err := dockerclient.NewDockerClient(config.DockerApiEndpoint, config.RegistryCredentials)
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

		// sleep for a few seconds just to let any active requests finish up
		// gracefully if possible. TODO: We should really only do this (same goes
		// for reloading the nginx config) if something has actually changed.
		logrus.Debug("Sleeping to allow active requests to finish gracefully")
		time.Sleep(5 * time.Second)

	}

	err = dockerClient.RemoveOldContainers(runningContainers)
	exitOnError(err, "Unable to stop and remove old containers")

}
