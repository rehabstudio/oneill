package dockerclient

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"

	"github.com/rehabstudio/oneill/config"
)

var (
	client      *docker.Client
	credentials map[string]docker.AuthConfiguration
)

func InitDockerClient(endpoint string, registryCredentials map[string]config.RegistryCredentials) error {

	// connect to the docker daemon and initialise a new API client
	var err error
	client, err = docker.NewClient(endpoint)
	if err != nil {
		return err
	}

	// initialise a docker.AuthConfiguration struct for each set of registry credentials
	credentials = make(map[string]docker.AuthConfiguration)
	for key, value := range registryCredentials {
		credentials[key] = docker.AuthConfiguration{
			Username: value.Username,
			Password: value.Password,
		}
	}

	return nil
}

// GetContainerByName searches for an existing container by name, returning an
// error if not found.
func GetContainerByName(name string) (docker.APIContainers, error) {

	// list all existing containers
	containers, err := ListContainers()
	if err != nil {
		return docker.APIContainers{}, err
	}

	// check all containers to see if one matching our name is present
	for _, c := range containers {
		if strings.TrimPrefix(c.Names[0], "/") == name {
			return c, nil
		}
	}

	return docker.APIContainers{}, fmt.Errorf("Container not found: %s", name)
}

// InspectContainer is a simple proxy function that exposes the method of the
// same name from the instantiated docker client instance.
func InspectContainer(s string) (*docker.Container, error) {
	return client.InspectContainer(s)
}

// InspectImage is a simple proxy function that exposes the method of the same
// name from the instantiated docker client instance.
func InspectImage(s string) (*docker.Image, error) {
	return client.InspectImage(s)
}

// ListContainers returns a slice containing all existing docker containers on
// the current host (running or otherwise).
func ListContainers() ([]docker.APIContainers, error) {
	return client.ListContainers(docker.ListContainersOptions{All: true})
}

// PullImage pulls the latest image for the given repotag from a remote
// registry. Credentials (if provided) are used in all requests to private
// registries.
func PullImage(repoTag string) error {

	logrus.WithFields(logrus.Fields{
		"repo_tag": repoTag,
	}).Debug("Pulling latest image from registry")

	// configuration options that get passed to client.PullImage
	repository, tag := docker.ParseRepositoryTag(repoTag)
	pullImageOptions := docker.PullImageOptions{Repository: repository, Tag: tag}

	// parse registry name from repository string
	var registrystr string
	parts := strings.Split(repository, "/")
	if len(parts) <= 1 {
		registrystr = ""
	}
	registrystr = parts[0]

	// pull image from registry
	return client.PullImage(pullImageOptions, credentials[registrystr])
}

// RemoveContainer removes a single existing container.
func RemoveContainer(c docker.APIContainers) error {

	logrus.WithFields(logrus.Fields{
		"container": strings.TrimPrefix(c.Names[0], "/"),
	}).Info("Removing docker container")

	// force removal of the container (remove even if it's running)
	// along with any volumes it owns.
	if err := client.RemoveContainer(docker.RemoveContainerOptions{
		ID: c.ID, RemoveVolumes: true, Force: true,
	}); err != nil {
		return err
	}

	return nil
}

// StartContainer creates and starts a new container for the given container
// definition. The name and port of the newly running container will be
// returned along with the definition.
func StartContainer(name string, repoTag string, env []string, dockerControlEnabled bool, portMapping map[int]int) error {

	logrus.WithFields(logrus.Fields{
		"container_name": name,
		"repo_tag":       repoTag,
	}).Info("Starting docker container")

	// configure docker socket mount if required
	var binds []string
	if dockerControlEnabled {
		binds = []string{"/var/run/docker.sock:/var/run/docker.sock"}
	} else {
		binds = []string{}
	}

	// convert portMapping map into the map[Port][]PortBinding that docker expects
	portBindings := portMappingToPortBindings(portMapping)
	// convert portMapping map into the map[Port]struct{} that docker expects
	exposedPorts := make(map[docker.Port]struct{})
	for _, internalPort := range portMapping {
		exposedPorts[docker.Port(fmt.Sprintf("%d/tcp", internalPort))] = struct{}{}
	}

	// if we've got any explicitly exposed ports then don't publish all ports
	// for the container
	var publishAllPorts bool
	if len(exposedPorts) > 0 {
		publishAllPorts = false
	} else {
		publishAllPorts = true
	}

	hostConfig := docker.HostConfig{PublishAllPorts: publishAllPorts, RestartPolicy: docker.RestartOnFailure(10), Binds: binds, PortBindings: portBindings}
	createContainerOptions := docker.CreateContainerOptions{
		Name:       name,
		Config:     &docker.Config{Image: repoTag, Env: env, ExposedPorts: exposedPorts},
		HostConfig: &hostConfig,
	}

	container, err := client.CreateContainer(createContainerOptions)
	if err != nil {
		return err
	}

	err = client.StartContainer(container.ID, &hostConfig)
	if err != nil {
		return err
	}

	return nil
}
