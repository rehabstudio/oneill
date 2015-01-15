package processors

import (
	"fmt"
	"strings"

	"github.com/fsouza/go-dockerclient"
	"github.com/rehabstudio/oneill/oneill"
)

func removeContainerOptions(container docker.APIContainers) docker.RemoveContainerOptions {
	return docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
		Force:         true,
	}
}

func containerShouldBeRunning(c docker.APIContainers, siteConfigs []*oneill.SiteConfig) bool {
	containerName := strings.TrimPrefix(c.Names[0], "/")
	for _, sc := range siteConfigs {
		if sc.Subdomain == containerName {
			// check that the container is running
			runningContainer := getContainerByID(c.ID)
			if !runningContainer.State.Running {
				oneill.LogDebug(fmt.Sprintf("Container present but exited: %s", containerName))
				return false
			}
			// check that the image running is the latest that's available locally
			availableImage := getImageByID(fmt.Sprintf("%s:%s", sc.Container, sc.Tag))
			if runningContainer.Image == availableImage.ID {
				return true
			} else {
				oneill.LogDebug(fmt.Sprintf("Container running but not up to date: %s", containerName))
				return false
			}
		}
	}
	return false
}

func RemoveContainers(siteConfigs []*oneill.SiteConfig) []*oneill.SiteConfig {
	oneill.LogInfo("## Removing unnecessary containers")

	for _, c := range oneill.ListContainers() {
		if !containerShouldBeRunning(c, siteConfigs) {
			err := oneill.DockerClient.RemoveContainer(removeContainerOptions(c))
			oneill.ExitOnError(err, "Unable to remove docker container")
			containerName := strings.TrimPrefix(c.Names[0], "/")
			oneill.LogInfo(fmt.Sprintf("Removed container: %s", containerName))
		}
	}
	return siteConfigs
}
