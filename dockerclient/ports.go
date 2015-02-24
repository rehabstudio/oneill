package dockerclient

import (
	"fmt"
	"strconv"

	"github.com/fsouza/go-dockerclient"
)

// portMappingToPortBindings converts a map[int]int to
// map[docker.Port][]docker.PortBinding so that we can pass it to the docker
// api in a format it expects
func portMappingToPortBindings(portMapping map[int]int) map[docker.Port][]docker.PortBinding {

	pb := make(map[docker.Port][]docker.PortBinding)
	for exposedPort, internalPort := range portMapping {
		portBindingSlice := []docker.PortBinding{
			docker.PortBinding{HostIP: "0.0.0.0", HostPort: strconv.Itoa(exposedPort)},
		}
		pb[docker.Port(fmt.Sprintf("%d/tcp", internalPort))] = portBindingSlice
		pb[docker.Port(fmt.Sprintf("%d/udp", internalPort))] = portBindingSlice
	}

	return pb
}

// check that the given port is exposed in the bindings extracted from a
// running container
func portInActiveBindings(port int, bindings []docker.PortBinding) bool {

	strPort := strconv.Itoa(port)
	for _, binding := range bindings {
		if strPort == binding.HostPort {
			return true
		}
	}

	return false
}

// PortsMatch checks if a running container's exposed ports (those bound to
// the host interface) match those defined in a the container definition.
func PortsMatch(definedPorts map[int]int, runningPorts map[docker.Port][]docker.PortBinding) bool {

	for _, protocol := range []string{"tcp", "udp"} {
		for exposedPort, internalPort := range definedPorts {
			var dPort docker.Port = docker.Port(fmt.Sprintf("%d/%s", internalPort, protocol))
			portBindings, portMapped := runningPorts[dPort]
			if !portMapped {
				return false
			}
			if !portInActiveBindings(exposedPort, portBindings) {
				return false
			}
		}
	}

	return true
}
