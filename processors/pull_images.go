package processors

import (
	"fmt"
	"strings"

	"github.com/fsouza/go-dockerclient"

	"github.com/rehabstudio/oneill/config"
	"github.com/rehabstudio/oneill/logger"
	"github.com/rehabstudio/oneill/oneill"
)

func getRegistryForContainer(container string) string {
	parts := strings.Split(container, "/")
	if len(parts) <= 1 {
		return ""
	}
	return parts[0]
}

func authConfiguration(siteConfig *oneill.SiteConfig) docker.AuthConfiguration {
	registry := getRegistryForContainer(siteConfig.Container)
	credentials := config.Config.RegistryCredentials[registry]
	return docker.AuthConfiguration{
		Username: credentials.Username,
		Password: credentials.Password,
	}
}

func pullImageOptions(siteConfig *oneill.SiteConfig) docker.PullImageOptions {
	return docker.PullImageOptions{
		Repository: siteConfig.Container,
		// Tag:        siteConfig.Tag,
	}
}

func PullImages(siteConfigs []*oneill.SiteConfig) []*oneill.SiteConfig {
	logger.LogInfo("## Pulling latest images from registry")

	for _, sc := range siteConfigs {
		logger.LogDebug(fmt.Sprintf("Pulling container image from registry: %s:%s", sc.Container, sc.Tag))
		err := oneill.DockerClient.PullImage(pullImageOptions(sc), authConfiguration(sc))
		if err != nil {
			logger.LogWarning(fmt.Sprintf("Unable to pull image from registry %s:%s", sc.Container, sc.Tag))
		}
	}
	return siteConfigs
}
