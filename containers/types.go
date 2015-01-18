package containers

import (
	"github.com/fsouza/go-dockerclient"

	"github.com/rehabstudio/oneill/definitions"
)

type Client interface {
	CheckOnlyOnePort(*docker.Image) bool
	ContainerRunning(string) bool
	ContainerShouldBeRunning(docker.APIContainers, map[string]*definitions.ContainerDefinition) bool
	GetContainerByName(string) (docker.APIContainers, error)
	GetPortFromContainer(docker.APIContainers) (int64, error)
	ListContainers() ([]docker.APIContainers, error)
	ListImages() ([]docker.APIImages, error)
	LoadImage(string, string) (*docker.Image, error)
	PullImage(string, string) error
	RemoveContainer(docker.APIContainers) error
	StartContainer(string, string, string) error
}

type ContainerDefinition struct {
	Subdomain string `yaml:"subdomain"`
	Container string `yaml:"container"`
	Tag       string `yaml:"tag"`
}
