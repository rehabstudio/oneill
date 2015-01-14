package processors

import (
	"fmt"
	"github.com/rehabstudio/oneill/oneill"
	"regexp"
)

var (
	rxContainerName = regexp.MustCompile(`^/?[a-zA-Z0-9_-]+$`)
)

func definitionIsUnique(siteConfig *oneill.SiteConfig, siteConfigs []*oneill.SiteConfig) bool {
	var count int
	for _, sc := range siteConfigs {
		if sc.Subdomain == siteConfig.Subdomain {
			count = count + 1
		}
	}
	return count <= 1
}

func validateSiteDefinition(siteConfig *oneill.SiteConfig, siteConfigs []*oneill.SiteConfig) bool {
	if !rxContainerName.MatchString(siteConfig.Subdomain) {
		oneill.LogWarning(fmt.Sprintf("%s is not a valid container name", siteConfig.Subdomain))
		return false
	}
	if !definitionIsUnique(siteConfig, siteConfigs) {
		oneill.LogWarning(fmt.Sprintf("%s is not unique", siteConfig.Subdomain))
		return false
	}
	oneill.LogDebug(fmt.Sprintf("%s is valid", siteConfig.Subdomain))
	return true
}

func ValidateDefinitionsPrePull(siteConfigs []*oneill.SiteConfig) []*oneill.SiteConfig {
	oneill.LogInfo("## Validating site definitions (pre pull)")

	var newSiteConfigs []*oneill.SiteConfig
	for _, siteConfig := range siteConfigs {
		oneill.LogDebug(fmt.Sprintf("Validating %s", siteConfig.Subdomain))
		if validateSiteDefinition(siteConfig, siteConfigs) {
			newSiteConfigs = append(newSiteConfigs, siteConfig)
		}
	}
	return newSiteConfigs
}

func ValidateDefinitionsPostPull(siteConfigs []*oneill.SiteConfig) []*oneill.SiteConfig {
	oneill.LogInfo("## Validating site definitions (post pull)")
	return siteConfigs
}
