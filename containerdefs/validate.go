package containerdefs

import (
	"fmt"
	"os"
	"regexp"

	"github.com/Sirupsen/logrus"
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
	logrus.WithFields(logrus.Fields{
		"definition": cd.Subdomain,
	}).Debug("Validating container definition")

	if len(cd.Subdomain) < 3 {
		logrus.WithFields(logrus.Fields{
			"definition": cd.Subdomain,
		}).Warning("subdomain not long enough (must be at least 3 characters long)")
		return false
	}
	if cd.Image == "" {
		logrus.WithFields(logrus.Fields{
			"definition": cd.Subdomain,
		}).Warning("image missing in container definition")
		return false
	}
	if !rxContainerName.MatchString(cd.Subdomain) {
		logrus.WithFields(logrus.Fields{
			"definition": cd.Subdomain,
		}).Warning("not a valid container name/subdomain")
		return false
	}
	if !definitionIsUnique(cd, cds) {
		logrus.WithFields(logrus.Fields{
			"definition": cd.Subdomain,
		}).Warning("subdomain must be unique")
		return false
	}

	return true
}
