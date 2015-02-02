package containerdefs

import (
	"fmt"
	"sync"
)

// processContainerDefinition processes an individual container definition,
// first pulling the image, validating it then starting a new container if
// necessary.
func processContainerDefinition(cdef *ContainerDefinition, cproc ContainerProcessor) (*RunningContainerDefinition, error) {

	// pull docker image if available (doesn't matter if not, we'll fail later)
	cproc.PullImage(cdef)

	// validate docker image (check it exists, exposes exactly 1 port, etc)
	err := cproc.ValidateImage(cdef)
	if err != nil {
		return &RunningContainerDefinition{}, err
	}

	// check if an already existing container matches the spec of the
	// container we want to start, if so then just return that one instead of starting a new one.
	runningContainerDefinition, err := cproc.GetExistingContainer(cdef)
	if err == nil {
		// found an existing container that matches our spec, so return it
		return runningContainerDefinition, nil
	}

	// create and start the container (if required)
	runningContainerDefinition, err = cproc.StartContainer(cdef)
	return runningContainerDefinition, err

}

// ProcessContainerDefinitions runs a goroutine for each definition, pulling
// images, validating them and starting containers if necessary.
func ProcessContainerDefinitions(cdefs []*ContainerDefinition, cproc ContainerProcessor) ([]*RunningContainerDefinition, error) {

	var wg sync.WaitGroup
	var rcdefs []*RunningContainerDefinition

	channel := make(chan interface{}, 1000)

	// process all container definitions concurrently, we wait for all
	// goroutines to finish processing before cleaning up and switching to the
	// newly generated configuration.
	for _, cdef := range cdefs {

		wg.Add(1)
		go func(cdef *ContainerDefinition) {
			defer wg.Done()
			runningContainerDefinition, err := processContainerDefinition(cdef, cproc)
			if err != nil {
				channel <- err
			} else {
				channel <- runningContainerDefinition
			}
		}(cdef)

	}
	// wait for all goroutines to complete
	wg.Wait()
	close(channel)

	for result := range channel {
		switch result.(type) {
		case *RunningContainerDefinition:
			rcdefs = append(rcdefs, result.(*RunningContainerDefinition))
		case error:
			// we'll silently do nothing here since the goroutine will have
			// already logged any warning messages required (we might at a
			// later point use this error value though, so we'll catch it anyway)
		default:
			return rcdefs, fmt.Errorf("Unexpected type from channel: %v", result)
		}
	}

	return rcdefs, nil
}
