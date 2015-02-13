package containerdefs

import (
	"sync"

	"github.com/Sirupsen/logrus"

	"github.com/rehabstudio/oneill/config"
	"github.com/rehabstudio/oneill/dockerclient"
)

// processContainerDefinition processes an individual container definition,
// first pulling the image, then starting a new container if necessary.
func processContainerDefinition(conf *config.Configuration, cd *ContainerDefinition) {

	// pull docker image if available (doesn't matter if not, we'll fail later)
	dockerclient.PullImage(cd.RepoTag)

	// check if an already existing container matches the spec of the
	// container we want to start, if so then we can stop processing this
	// definition.
	if cd.AlreadyRunning(conf.PersistenceDirectory) {
		logrus.WithFields(logrus.Fields{
			"container_name": cd.ContainerName,
		}).Debug("Container already running, no action taken")
		return
	}

	// remove container if one is running with the same name since we know
	// it's not configured correctly (or we would have bailed out by now)
	err := cd.RemoveContainer()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"container_name": cd.ContainerName,
			"err":            err,
		}).Error("Unable to remove docker container")
		return
	}

	// create and start the new container
	err = cd.StartContainer(conf.PersistenceDirectory)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"container_name": cd.ContainerName,
			"err":            err,
		}).Error("Unable to start docker container")
		return
	}
}

// ProcessContainerDefinitions runs a goroutine for each definition, pulling
// images, validating them and starting containers if necessary.
func ProcessContainerDefinitions(conf *config.Configuration, cdefs []*ContainerDefinition) error {

	// process all container definitions concurrently
	var wg sync.WaitGroup
	for _, cdef := range cdefs {
		wg.Add(1)
		go func(cdef *ContainerDefinition) {
			defer wg.Done()
			processContainerDefinition(conf, cdef)
		}(cdef)
	}

	// wait for all goroutines to complete before returning
	wg.Wait()
	return nil
}
