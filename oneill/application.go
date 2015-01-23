package oneill

import (
	"fmt"

	"github.com/rehabstudio/oneill/config"
	"github.com/rehabstudio/oneill/containers"
	"github.com/rehabstudio/oneill/definitions"
	"github.com/rehabstudio/oneill/logger"
	"github.com/rehabstudio/oneill/proxy"
)

type Application struct {
	config               *config.Configuration
	dockerClient         containers.Client
	containerDefinitions map[string]*definitions.ContainerDefinition
}

func NewApplication(c *config.Configuration, dc containers.Client, cds []*definitions.ContainerDefinition) *Application {
	logger.L.Debug("Initialising oneill instance")

	// initialise new application struct
	app := Application{
		config:               c,
		dockerClient:         dc,
		containerDefinitions: make(map[string]*definitions.ContainerDefinition),
	}
	for _, cd := range cds {
		app.containerDefinitions[cd.Subdomain] = cd
	}

	return &app
}

func (a *Application) RunApplication() error {

	// pull latest docker image/tag for each container definition, we don't
	// *really* care if this passes or fails, so long as at the next step
	// there's at least one container matching the definition
	logger.L.Debug("Pulling latest images for all configured containers")
	for _, cd := range a.containerDefinitions {
		a.dockerClient.PullImage(cd.Image, cd.Tag)
	}

	// for each container definition validate the docker image that accompanies it
	// If either of these checks fail, the container definition is removed from
	//processing for any following steps.
	for k, cd := range a.containerDefinitions {
		// check that an image with the appropriate tag exists locally (should
		// have been pulled in the last step)
		image, err := a.dockerClient.LoadImage(cd.Image, cd.Tag)
		if err != nil {
			logger.L.Warning(fmt.Sprintf("Unable to find image, skipping: %s (%s:%s)", cd.Subdomain, cd.Image, cd.Tag))
			delete(a.containerDefinitions, k)
			continue
		}
		// check that the image exposes exactly 1 port.
		if !a.dockerClient.CheckOnlyOnePort(image) {
			logger.L.Warning(fmt.Sprintf("Image does not expose a single port, skipping: %s (%s:%s)", cd.Subdomain, cd.Image, cd.Tag))
			delete(a.containerDefinitions, k)
			continue
		}
	}

	// iterate over all active docker containers, ensuring that each one matches
	// a currently valid container definition. If no matching container definition
	// is found, the container will be stopped and forcibly removed. This means
	// that until (if) persistent volume support is implemented any docker containers
	// created by oneill will only have ephemeral filesystem storage.
	logger.L.Debug("Removing containers that don't match any valid definition")
	containers, err := a.dockerClient.ListContainers()
	if err != nil {
		return err
	}
	for _, container := range containers {
		if !a.dockerClient.ContainerShouldBeRunning(container, a.containerDefinitions) {
			a.dockerClient.RemoveContainer(container)
		}
	}

	// iterate over all active container definitions, ensuring that a container is started for each one.
	logger.L.Debug("Ensure a container is running for each valid definition")
	for _, cd := range a.containerDefinitions {
		// if a container is already running for the given name then skip forward to the next definition
		if a.dockerClient.ContainerRunning(cd.Subdomain) {
			continue
		}
		// create and start the container
		a.dockerClient.StartContainer(cd.Subdomain, cd.Image, cd.Tag, cd.Env)
	}

	// Clear out any existing files in the directory
	logger.L.Debug("Removing all existing reverse proxy configuration")
	err = proxy.ClearConfigDirectory(a.config.NginxConfigDirectory)
	// exit if this fails because it means we probably can't manage
	// the directory, so we won't try
	if err != nil {
		return err
	}

	// write nginx templates to disk in the configured folder.
	for _, cd := range a.containerDefinitions {
		// if a container isn't running then don't configure it
		if !a.dockerClient.ContainerRunning(cd.Subdomain) {
			continue
		}
		// grab currently exposed port from the running container
		container, err := a.dockerClient.GetContainerByName(cd.Subdomain)
		if err != nil {
			continue
		}
		// for some reason occasionally docker fails to map the port correctly when starting
		// the container, so it's important we check before attempting to write the configuration
		port, err := a.dockerClient.GetPortFromContainer(container)
		if err != nil {
			continue
		}
		// write proxy (nginx) configuration file to disk
		proxy.WriteConfig(a.config.NginxConfigDirectory, a.config.NginxHtpasswdDirectory, a.config.ServingDomain, cd.Subdomain, cd.Htpasswd, port, a.config.NginxSSLEnabled, a.config.NginxSSLCertPath, a.config.NginxSSLKeyPath)
	}

	// finally, reload the proxy server by sending a HUP signal, this performs a hotreload without
	// any downtime due to configuration loading
	logger.L.Debug("Reloading reverse proxy configuration")
	err = proxy.ReloadServer()
	if err != nil {
		logger.L.Warning("Unable to reload nginx configuration")
	}

	return nil
}
