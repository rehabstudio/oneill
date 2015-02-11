package containerdefs

import (
	"strings"

	"github.com/rehabstudio/oneill/dockerclient"
)

func nameInContainerDefs(name string, cdefs []*ContainerDefinition) bool {

	for _, cdef := range cdefs {
		if name == cdef.ContainerName {
			return true
		}
	}

	return false
}

// RemoveRedundantContainers loops through all running docker containers and
// stops/removes any that whose name doesn't match the name of one of the
// container definitions passed into the function.
func RemoveRedundantContainers(cdefs []*ContainerDefinition) error {

	containers, err := dockerclient.ListContainers()
	if err != nil {
		return err
	}

	for _, c := range containers {
		cName := strings.TrimPrefix(c.Names[0], "/")
		if !nameInContainerDefs(cName, cdefs) {
			err := dockerclient.RemoveContainer(c)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
