package dockerclient

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"

	"github.com/rehabstudio/oneill/config"
	"github.com/rehabstudio/oneill/containerdefs"
)

type dockerClient struct {
	endpoint    string
	client      *docker.Client
	credentials map[string]docker.AuthConfiguration
}

// NewDockerClient returns a new docker client, preconfigured with the
// credentials for any private registry we might want to use.
func NewDockerClient(endpoint string, registryCredentials map[string]config.RegistryCredentials) (*dockerClient, error) {

	// connect to the docker daemon and initialise a new API client
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return &dockerClient{}, err
	}

	// initialise docker client
	dc := dockerClient{
		endpoint: endpoint,
		client:   client,
	}

	// initialise a docker.AuthConfiguration struct for each set of registry credentials
	dc.credentials = make(map[string]docker.AuthConfiguration)
	for key, value := range registryCredentials {
		dc.credentials[key] = docker.AuthConfiguration{
			Username: value.Username,
			Password: value.Password,
		}
	}

	return &dc, nil
}

// envsMatch checks if a running container's environment matches the one
// defined in a container definition. The variables defined in the container
// definition are added to those defined in the base image before comparing
// with those read from the running container.
func envsMatch(env0, env1, fromImage []string) bool {

	for _, v := range fromImage {
		env0 = append(env0, v)
	}

	if len(env0) != len(env1) {
		return false
	}

	sort.Strings(env0)
	sort.Strings(env1)
	for i := range env0 {
		if env0[i] != env1[i] {
			return false
		}
	}

	return true
}

// getContainerByName searches for an existing container by name, returning an
// error if not found.
func (d *dockerClient) getContainerByName(name string) (docker.APIContainers, error) {

	// list all existing containers
	containers, err := d.ListContainers()
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

// getPortFromContainer looks up a container by name and returns the currently
// exposed port. An error is returned if the container is not exposing any
// ports.
func (d *dockerClient) getPortFromContainer(cName string) (int64, error) {

	c, err := d.getContainerByName(cName)
	if err != nil {
		return 0, err
	}

	if len(c.Ports) < 1 {
		cName := strings.TrimPrefix(c.Names[0], "/")
		err := fmt.Errorf("Container not exposing any ports: %s", cName)
		return 0, err
	}

	return c.Ports[0].PublicPort, nil
}

// ListContainers returns a slice containing all existing docker containers on
// the current host (running or otherwise).
func (d *dockerClient) ListContainers() ([]docker.APIContainers, error) {
	return d.client.ListContainers(docker.ListContainersOptions{All: true})
}

// parseRegistryFromImageName parses a "repotag" and returns the name of the
// private registry being used. This is useful when looking up which
// credentials to use when pulling an image.
func parseRegistryFromImageName(image string) string {
	parts := strings.Split(image, "/")
	if len(parts) <= 1 {
		return ""
	}
	return parts[0]
}

// PullImage pulls the latest image for the given container definition from a
// remote registry. Credentials (if provided) are used in all requests to
// private registries.
func (d *dockerClient) PullImage(cd *containerdefs.ContainerDefinition) error {
	logrus.WithFields(logrus.Fields{
		"image": cd.Image,
		"tag":   cd.Tag,
	}).Debug("Pulling docker image")

	// configuration options that get passed to client.PullImage
	pullImageOptions := docker.PullImageOptions{Repository: cd.Image, Tag: cd.Tag}

	// auth configuration struct for the registry being used
	registry := parseRegistryFromImageName(cd.Image)
	authConfiguration := d.credentials[registry]

	return d.client.PullImage(pullImageOptions, authConfiguration)
}

// containerInRunningDefinitions checks if a given container has a matching
// (running) container definition.
func containerInRunningDefinitions(c docker.APIContainers, rcds []*containerdefs.RunningContainerDefinition) bool {
	containerName := strings.TrimPrefix(c.Names[0], "/")
	for _, rcd := range rcds {
		if containerName == rcd.Name {
			return true
		}
	}
	return false
}

// RemoveOldContainers removes any existing container that does not have a
// match (by name) in the passed in slice of currently running container
// definitions.
func (d *dockerClient) RemoveOldContainers(rcds []*containerdefs.RunningContainerDefinition) error {

	// get list of currently existing containers
	containerList, err := d.ListContainers()
	if err != nil {
		return err
	}

	// loop over all containers, stopping and removing any that shouldn't be
	// there.
	for _, c := range containerList {
		if !containerInRunningDefinitions(c, rcds) {

			logrus.WithFields(logrus.Fields{
				"container": strings.TrimPrefix(c.Names[0], "/"),
			}).Info("Removing docker container")

			// force removal of the container (remove even if it's running)
			// along with any volumes it owns.
			if err := d.client.RemoveContainer(docker.RemoveContainerOptions{
				ID: c.ID, RemoveVolumes: true, Force: true,
			}); err != nil {
				return err
			}

		}
	}

	return nil
}

// GetExistingContainer searches for a currently existing container that
// matches exactly the given container definition. In addition to verifying
// the data matches the definition, it checks that the container is both
// running and exposing a port.
func (d *dockerClient) GetExistingContainer(cd *containerdefs.ContainerDefinition) (*containerdefs.RunningContainerDefinition, error) {

	containerPrefix := fmt.Sprintf("oneill-%s-", cd.Subdomain)

	containerList, err := d.ListContainers()
	if err != nil {
		return &containerdefs.RunningContainerDefinition{}, err
	}

	// check that an image with the given tag actually still exists (this is
	// probably a bit paranoid, but performance isn't critical here)
	availableImage, err := d.client.InspectImage(fmt.Sprintf("%s:%s", cd.Image, cd.Tag))
	if err != nil {
		return &containerdefs.RunningContainerDefinition{}, err
	}

	for _, c := range containerList {

		// skip images that don't match this container's prefix/name
		containerName := strings.TrimPrefix(c.Names[0], "/")
		if !strings.HasPrefix(containerName, containerPrefix) {
			continue
		}

		// check that the container is actually running
		runningContainer, err := d.client.InspectContainer(c.ID)
		if err != nil {
			continue
		}
		if !runningContainer.State.Running {
			continue
		}

		// check that the image running is the latest that's available locally
		if runningContainer.Image != availableImage.ID {
			continue
		}

		// check that the running container's environment matches the one in
		// the container definition
		if !envsMatch(cd.Env, runningContainer.Config.Env, availableImage.Config.Env) {
			continue
		}

		port, err := d.getPortFromContainer(containerName)
		if err != nil {
			return &containerdefs.RunningContainerDefinition{}, err
		}

		rcd := containerdefs.RunningContainerDefinition{
			ContainerDefinition: cd,
			Name:                containerName,
			Port:                port,
		}
		return &rcd, nil

	}

	err = fmt.Errorf("Unable to find running container for definition: %s", cd.Subdomain)
	return &containerdefs.RunningContainerDefinition{}, err
}

// StartContainer creates and starts a new container for the given container
// definition. The name and port of the newly running container will be
// returned along with the definition.
func (d *dockerClient) StartContainer(cd *containerdefs.ContainerDefinition) (*containerdefs.RunningContainerDefinition, error) {
	logrus.WithFields(logrus.Fields{
		"subdomain": cd.Subdomain,
		"image":     cd.Image,
		"tag":       cd.Tag,
	}).Info("Starting docker container")

	containerName := fmt.Sprintf("oneill-%s-%d", cd.Subdomain, time.Now().Unix())
	hostConfig := docker.HostConfig{PublishAllPorts: true, RestartPolicy: docker.RestartOnFailure(10)}
	createContainerOptions := docker.CreateContainerOptions{
		Name:       containerName,
		Config:     &docker.Config{Image: fmt.Sprintf("%s:%s", cd.Image, cd.Tag), Env: cd.Env},
		HostConfig: &hostConfig,
	}

	container, err := d.client.CreateContainer(createContainerOptions)
	if err != nil {
		return &containerdefs.RunningContainerDefinition{}, err
	}

	err = d.client.StartContainer(container.ID, &hostConfig)
	if err != nil {
		return &containerdefs.RunningContainerDefinition{}, err
	}

	port, err := d.getPortFromContainer(container.Name)
	if err != nil {
		return &containerdefs.RunningContainerDefinition{}, err
	}

	rcd := containerdefs.RunningContainerDefinition{
		ContainerDefinition: cd,
		Name:                containerName,
		Port:                port,
	}
	return &rcd, nil
}

// ValidateImage performs several checks against a downloaded docker image and
// ensures it's suitable for use.
func (d *dockerClient) ValidateImage(cd *containerdefs.ContainerDefinition) error {
	logrus.WithFields(logrus.Fields{
		"image": cd.Image,
		"tag":   cd.Tag,
	}).Debug("Validating docker image")

	// check that an image with the appropriate tag exists locally (should
	// have been pulled in the last step)
	image, err := d.client.InspectImage(fmt.Sprintf("%s:%s", cd.Image, cd.Tag))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"image": cd.Image,
			"tag":   cd.Tag,
			"err":   err,
		}).Warning("Unable to find image, skipping")
		return err
	}

	// check that the image exposes exactly 1 port. For now oneill doesn't
	// support containers unless they expose exactly one port, this is to make
	// configuration and interaction with nginx much simpler. We may revise
	// this decision in future.
	if len(image.Config.ExposedPorts) != 1 {
		errStr := "Image does not expose a single port, skipping"
		err = fmt.Errorf(errStr)
		logrus.WithFields(logrus.Fields{
			"image": cd.Image,
			"tag":   cd.Tag,
			"err":   err,
		}).Warning(errStr)
		return err
	}

	return nil
}
