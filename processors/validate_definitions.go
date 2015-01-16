package processors

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/fsouza/go-dockerclient"

	"github.com/rehabstudio/oneill/logger"
	"github.com/rehabstudio/oneill/oneill"
)

var (
	rxContainerName = regexp.MustCompile(`^/?[a-zA-Z0-9_-]+$`)
)

// ByAge implements sort.Interface for []Person based on
// the Age field.
type ByCreated []docker.APIImages

func (a ByCreated) Len() int           { return len(a) }
func (a ByCreated) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreated) Less(i, j int) bool { return a[i].Created > a[j].Created }

func checkOnlyOneExposedPort(image *docker.Image) bool {
	return len(image.Config.ExposedPorts) == 1
}

func definitionIsUnique(siteConfig *oneill.SiteConfig, siteConfigs []*oneill.SiteConfig) bool {
	var count int
	for _, sc := range siteConfigs {
		if sc.Subdomain == siteConfig.Subdomain {
			count = count + 1
		}
	}
	return count <= 1
}

func getContainerForSiteConfig(siteConfig *oneill.SiteConfig, containers []docker.APIContainers) (docker.APIContainers, error) {
	for _, c := range containers {
		containerName := strings.TrimPrefix(c.Names[0], "/")
		if containerName == siteConfig.Subdomain {
			return c, nil
		}
	}
	return docker.APIContainers{}, fmt.Errorf("Container not found for %s", siteConfig.Subdomain)
}

func getContainerByID(cid string) *docker.Container {
	container, _ := oneill.DockerClient.InspectContainer(cid)
	return container
}

func getImageByID(iid string) *docker.Image {
	image, _ := oneill.DockerClient.InspectImage(iid)
	return image
}

func getImageForSiteConfig(siteConfig *oneill.SiteConfig, images []docker.APIImages) (docker.APIImages, error) {
	sort.Sort(ByCreated(images))
	for _, i := range images {
		siteConfigRepoTag := fmt.Sprintf("%s:%s", siteConfig.Container, siteConfig.Tag)
		for _, rt := range i.RepoTags {
			if rt == siteConfigRepoTag {
				return i, nil
			}
		}
	}
	return docker.APIImages{}, fmt.Errorf("Image not found for %s:%s", siteConfig.Container, siteConfig.Tag)
}

func validateSiteDefinition(siteConfig *oneill.SiteConfig, siteConfigs []*oneill.SiteConfig) bool {
	if !rxContainerName.MatchString(siteConfig.Subdomain) {
		logger.LogWarning(fmt.Sprintf("%s is not a valid container name", siteConfig.Subdomain))
		return false
	}
	if !definitionIsUnique(siteConfig, siteConfigs) {
		logger.LogWarning(fmt.Sprintf("%s is not unique", siteConfig.Subdomain))
		return false
	}
	logger.LogDebug(fmt.Sprintf("%s is valid", siteConfig.Subdomain))
	return true
}

func ValidateDefinitionsPrePull(siteConfigs []*oneill.SiteConfig) []*oneill.SiteConfig {
	logger.LogInfo("## Validating site definitions (pre pull)")

	var newSiteConfigs []*oneill.SiteConfig
	for _, siteConfig := range siteConfigs {
		logger.LogDebug(fmt.Sprintf("Validating %s", siteConfig.Subdomain))
		if validateSiteDefinition(siteConfig, siteConfigs) {
			newSiteConfigs = append(newSiteConfigs, siteConfig)
		}
	}
	return newSiteConfigs
}

func ValidateDefinitionsPostPull(siteConfigs []*oneill.SiteConfig) []*oneill.SiteConfig {
	logger.LogInfo("## Validating site definitions (post pull)")

	images := oneill.ListImages()
	var newSiteConfigs []*oneill.SiteConfig
	for _, siteConfig := range siteConfigs {
		// check that an appropriate image exists locally
		apiImage, err := getImageForSiteConfig(siteConfig, images)
		if err != nil {
			logger.LogWarning(fmt.Sprintf("%s:%s cannot be found", siteConfig.Container, siteConfig.Tag))
			continue
		}
		// check that the image only exposes one port
		image := getImageByID(apiImage.ID)
		if !checkOnlyOneExposedPort(image) {
			logger.LogWarning(fmt.Sprintf("%s:%s does not expose exactly 1 port", siteConfig.Container, siteConfig.Tag))
			continue
		}
		newSiteConfigs = append(newSiteConfigs, siteConfig)
	}
	return newSiteConfigs
}
