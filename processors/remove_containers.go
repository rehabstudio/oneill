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

func RemoveContainers(siteConfigs []*oneill.SiteConfig) []*oneill.SiteConfig {
	oneill.LogInfo("## Removing unnecessary containers")

	for _, c := range oneill.ListContainers() {
		containerName := strings.TrimPrefix(c.Names[0], "/")
		if !oneill.ContainerShouldBeRunning(containerName, siteConfigs) {
			err := oneill.DockerClient.RemoveContainer(removeContainerOptions(c))
			if err != nil {
				panic(err)
			}
			oneill.LogInfo(fmt.Sprintf("Removed container: %s", containerName))
		}
	}
	return siteConfigs
}
