package definitions

import (
	"fmt"
	"os"
	"regexp"

	"github.com/rehabstudio/oneill/logger"
)

var (
	rxContainerName = regexp.MustCompile(`^/?[a-zA-Z0-9_-]+$`)
)

func isDirectory(path string) error {

	// check if rootDirectory exists
	src, err := os.Stat(path)
	if err != nil {
		return err
	}

	// check if rootDirectory is actually a directory
	if !src.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	return nil
}

func definitionIsUnique(cd *ContainerDefinition, cds []*ContainerDefinition) bool {
	var count int
	for _, ocd := range cds {
		if ocd.Subdomain == cd.Subdomain {
			count = count + 1
		}
	}
	return count <= 1
}

func validateDefinition(cd *ContainerDefinition, cds []*ContainerDefinition) bool {
	logger.L.Debug(fmt.Sprintf("Validating container definition: %s", cd.Subdomain))

	if len(cd.Subdomain) < 3 {
		logger.L.Warning(fmt.Sprintf("%s is not long enough (must be at least 3 characters long)", cd.Subdomain))
		return false
	}
	if cd.Image == "" {
		logger.L.Warning(fmt.Sprintf("`Image` cannot be blank in container definition: %s", cd.Subdomain))
		return false
	}
	if !rxContainerName.MatchString(cd.Subdomain) {
		logger.L.Warning(fmt.Sprintf("%s is not a valid container name", cd.Subdomain))
		return false
	}
	if !definitionIsUnique(cd, cds) {
		logger.L.Warning(fmt.Sprintf("%s is not unique", cd.Subdomain))
		return false
	}

	return true
}
