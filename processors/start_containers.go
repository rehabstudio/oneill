package processors

import (
	"fmt"

	"github.com/fsouza/go-dockerclient"

	"github.com/rehabstudio/oneill/logger"
	"github.com/rehabstudio/oneill/oneill"
)

func StartContainers(siteConfigs []*oneill.SiteConfig) []*oneill.SiteConfig {
	logger.LogInfo("## Starting required containers")
	activeContainers := oneill.ListContainers()

	for _, sc := range siteConfigs {
		if oneill.ContainerIsRunning(sc.Subdomain, activeContainers) {
			continue
		}
		config := docker.Config{Image: sc.Container}
		hostConfig := docker.HostConfig{PublishAllPorts: true}
		container, err := oneill.DockerClient.CreateContainer(docker.CreateContainerOptions{
			Name:       sc.Subdomain,
			Config:     &config,
			HostConfig: &hostConfig,
		})
		if err != nil {
			continue
		}
		err = oneill.DockerClient.StartContainer(container.ID, &hostConfig)
		if err != nil {
			continue
		}
		logger.LogInfo(fmt.Sprintf("Started container: %s", sc.Subdomain))
	}
	return siteConfigs
}
