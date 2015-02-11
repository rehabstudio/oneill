package dockerclient

// DockerSocketMounted checks that the unix socket docker uses to expose its
// API has been bind-mounted into the container.
func DockerSocketMounted(binds []string) bool {

	for _, bind := range binds {
		if bind == "/var/run/docker.sock:/var/run/docker.sock" {
			return true
		}
	}

	return false
}
