package containers

import (
	"fmt"
	"strings"

	"github.com/fsouza/go-dockerclient"

	"github.com/rehabstudio/oneill/config"
	"github.com/rehabstudio/oneill/definitions"
	"github.com/rehabstudio/oneill/logger"
)

type dockerClient struct {
	endpoint    string
	client      *docker.Client
	credentials map[string]docker.AuthConfiguration
}

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

func (d *dockerClient) CheckOnlyOnePort(image *docker.Image) bool {
	return len(image.Config.ExposedPorts) == 1
}

func (d *dockerClient) ContainerRunning(name string) bool {
	_, err := d.GetContainerByName(name)
	if err != nil {
		return false
	}
	return true
}

func (d *dockerClient) ContainerShouldBeRunning(container docker.APIContainers, definitions map[string]*definitions.ContainerDefinition) bool {

	// check that a definition with the container's name exists
	containerName := strings.TrimPrefix(container.Names[0], "/")
	definition, ok := definitions[containerName]
	if !ok {
		return false
	}

	// check that the container is actually running
	runningContainer := d.getContainerByID(container.ID)
	if !runningContainer.State.Running {
		return false
	}

	// check that an image with the given tag actually still exists (this is
	// probably a bit paranoid, but performance isn't critical here)
	availableImage, err := d.LoadImage(definition.Image, definition.Tag)
	if err != nil {
		return false
	}

	// check that the image running is the latest that's available locally
	if runningContainer.Image != availableImage.ID {
		// container running but out-of date
		return false
	}

	return true
}

func (d *dockerClient) getContainerByID(cid string) *docker.Container {
	container, _ := d.client.InspectContainer(cid)
	return container
}

func (d *dockerClient) GetContainerByName(name string) (docker.APIContainers, error) {

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

func (d *dockerClient) GetPortFromContainer(c docker.APIContainers) (int64, error) {
	if len(c.Ports) < 1 {
		cName := strings.TrimPrefix(c.Names[0], "/")
		err := fmt.Errorf("Container not exposing any ports: %s", cName)
		return 0, err
	}
	return c.Ports[0].PublicPort, nil
}

func (d *dockerClient) ListContainers() ([]docker.APIContainers, error) {
	return d.client.ListContainers(docker.ListContainersOptions{All: true})
}

func (d *dockerClient) ListImages() ([]docker.APIImages, error) {
	return d.client.ListImages(docker.ListImagesOptions{All: true})
}

func (d *dockerClient) LoadImage(image string, tag string) (*docker.Image, error) {

	dockerImage, err := d.client.InspectImage(fmt.Sprintf("%s:%s", image, tag))
	if err != nil {
		return &docker.Image{}, err
	}

	return dockerImage, nil
}

func parseRegistryFromImageName(image string) string {
	parts := strings.Split(image, "/")
	if len(parts) <= 1 {
		return ""
	}
	return parts[0]
}

func (d *dockerClient) PullImage(image string, tag string) error {
	logger.L.Debug(fmt.Sprintf("Pulling docker image: %s:%s", image, tag))

	// configuration options that get passed to client.PullImage
	pullImageOptions := docker.PullImageOptions{Repository: image, Tag: tag}

	// auth configuration struct for the registry being used
	registry := parseRegistryFromImageName(image)
	authConfiguration := d.credentials[registry]

	return d.client.PullImage(pullImageOptions, authConfiguration)
}

func (d *dockerClient) RemoveContainer(container docker.APIContainers) error {
	logger.L.Info(fmt.Sprintf("Removing docker container: %s", strings.TrimPrefix(container.Names[0], "/")))

	removeContainerOptions := docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
		Force:         true,
	}

	return d.client.RemoveContainer(removeContainerOptions)
}

func (d *dockerClient) StartContainer(subdomain, image, tag string) error {
	logger.L.Info(fmt.Sprintf("Starting docker container: %s (%s:%s)", subdomain, image, tag))

	hostConfig := docker.HostConfig{PublishAllPorts: true}
	createContainerOptions := docker.CreateContainerOptions{
		Name:       subdomain,
		Config:     &docker.Config{Image: fmt.Sprintf("%s:%s", image, tag)},
		HostConfig: &hostConfig,
	}

	container, err := d.client.CreateContainer(createContainerOptions)
	if err != nil {
		return err
	}

	return d.client.StartContainer(container.ID, &hostConfig)
}
