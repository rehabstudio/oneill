package oneill

import (
	"fmt"
	"os"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

const (
	dockerEndpoint string = "unix:///var/run/docker.sock"
)

var DockerClient *docker.Client

func InitDockerClient() {
	// connect to the docker daemon and initialise a new API client
	var err error
	DockerClient, err = docker.NewClient(dockerEndpoint)
	if err != nil {
		LogError(fmt.Sprintf("Error initialising docker client: %s", err))
		os.Exit(1)
	}
}

func ContainerIsRunning(name string, containers []docker.APIContainers) bool {
	for _, c := range containers {
		if strings.TrimPrefix(c.Names[0], "/") == name {
			return true
		}
	}
	return false
}

func ListContainers() []docker.APIContainers {
	c, err := DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	// if at any point we can't list containers we really don't want to continue
	if err != nil {
		LogError(fmt.Sprintf("Error listing local docker containers: %s", err))
		os.Exit(1)
	}
	return c
}

func ListImages() []docker.APIImages {
	c, err := DockerClient.ListImages(docker.ListImagesOptions{All: true})
	// if at any point we can't list images we really don't want to continue
	if err != nil {
		LogError(fmt.Sprintf("Error listing local docker images: %s", err))
		os.Exit(1)
	}
	return c
}
