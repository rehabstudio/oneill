package containerdefs

import (
	"regexp"

	"github.com/Sirupsen/logrus"

	"github.com/rehabstudio/oneill/dockerclient"
)

var (
	rxContainerName = regexp.MustCompile(`^/?[a-zA-Z0-9_-]+$`)
)

type ContainerDefinition struct {
	// ContainerName controls the user-specified part of the name oneill will
	// give to the container at startup time. oneill uses a simple convention
	// when naming containers `{prefix}-{container-name}`.
	// `prefix` specifies whether this container is an application or service
	// container, and `container-name` is specified by the user in a container
	// definition.
	ContainerName string `yaml:"container_name"`

	// RepoTag controls the container that will be pulled and run for this
	// container definition. This is in the same format as you would pass to
	// `docker run`, e.g. `locahost:5000/myimage:latest`, `nginx`,
	// `ubuntu:14.04`, `my.private.repo/myotherimage`
	RepoTag string `yaml:"repo_tag"`

	// Env is a slice containing arbitrary environment variables that get
	// passed to new containers at runtime. Variables set here will override
	// environment variables set anywhere else (including those set by oneill
	// itself), so use with caution.
	Env dockerclient.Env `yaml:"env"`

	// should persistence be enabled for this container? default off as we
	// don't want to encourage people to use persistence (whilst acknowledging
	// that it is necessary in some situations).
	PersistenceEnabled bool `yaml:"persistence_enabled"`

	// should the docker control socket be bind-mounted into this container?
	// this is useful for service containers that need to be able to see or
	// control what other containers are doing (nginx service, fluentd
	// service, etc. need this functionality).
	DockerControlEnabled bool `yaml:"docker_control_enabled"`

	// service containers allow an explicit port mapping as some services need
	// to be exposed on specific ports to be useful e.g. nginx on 80/443 for
	// serving http. Regular containers do not need this functionality. Keys
	// are host port numbers and values are the internal port numbers that
	// should be exposed.
	PortMapping map[int]int `yaml:"port_mapping"`
}

// AlreadyRunning checks whether a container is already running that matches
// *exactly* this container definition.
func (cd *ContainerDefinition) AlreadyRunning(persistenceDir string) bool {

	// check that an image with the given tag actually exists (container can't
	// be running if the image isn't there)
	availableImage, err := dockerclient.InspectImage(cd.RepoTag)
	if err != nil {
		return false
	}

	// grab an APIContainer by name
	c, err := dockerclient.GetContainerByName(cd.ContainerName)
	if err != nil {
		return false
	}

	// check that the container is actually running
	runningContainer, err := dockerclient.InspectContainer(c.ID)
	if err != nil {
		return false
	}
	if !runningContainer.State.Running {
		return false
	}

	// check that the image running is the latest that's available locally
	if runningContainer.Image != availableImage.ID {
		return false
	}

	// check that the running container's environment matches the one in
	// the container definition
	if !dockerclient.EnvsMatch(cd.Env, runningContainer.Config.Env, availableImage.Config.Env) {
		return false
	}

	// check that the running container has correctly bind-mounted the docker
	// socket (if configured to do so)
	if cd.DockerControlEnabled != dockerclient.DockerSocketMounted(runningContainer.HostConfig.Binds) {
		return false
	}

	// check that the running container has correctly bind-mounted the docker
	// containers directory (if configured to do so)
	if cd.DockerControlEnabled != dockerclient.DockerContainersDirMounted(runningContainer.HostConfig.Binds) {
		return false
	}

	// check that the running container's port mappings match those in the
	// container definition
	if !dockerclient.PortsMatch(cd.PortMapping, runningContainer.HostConfig.PortBindings) {
		return false
	}

	// check that the running container has correctly bind-mounted all volumes
	// if persistence is enabled in the definition.
	if cd.PersistenceEnabled && !dockerclient.AllVolumesMounted(cd.ContainerName, persistenceDir, runningContainer.Image, runningContainer.Volumes) {
		return false
	}

	return true
}

// RemoveContainer removes a container with the same name as contained within
// this definition if one exists.
func (cd *ContainerDefinition) RemoveContainer() error {

	container, err := dockerclient.GetContainerByName(cd.ContainerName)
	if err != nil {
		return nil
	}

	err = dockerclient.RemoveContainer(container)
	if err != nil {
		return err
	}

	return nil
}

// StartContainer assembles the appropriate options structs and starts a new
// container that matches the container definition.
func (cd *ContainerDefinition) StartContainer(persistenceDir string) error {
	return dockerclient.StartContainer(cd.ContainerName, cd.RepoTag, cd.Env, cd.DockerControlEnabled, cd.PersistenceEnabled, cd.PortMapping, persistenceDir)
}

// Validate checks that a container definition is internally consistent and
// that its configuration is valid in isolation. Validation of container
// definitions as a whole group happens (e.g. testing for uniqueness of
// container names or port mappings) elsewhere in the app.
func (cd *ContainerDefinition) Validate() bool {
	if len(cd.ContainerName) < 3 {
		logrus.WithFields(logrus.Fields{
			"container_name": cd.ContainerName,
		}).Warning("container_name not long enough (must be at least 3 characters long)")
		return false
	}

	if cd.RepoTag == "" {
		logrus.WithFields(logrus.Fields{
			"container_name": cd.ContainerName,
		}).Warning("repo_tag missing in container definition")
		return false
	}

	if !rxContainerName.MatchString(cd.ContainerName) {
		logrus.WithFields(logrus.Fields{
			"container_name": cd.ContainerName,
		}).Warning("not a valid value for container_name")
		return false
	}

	return true
}
