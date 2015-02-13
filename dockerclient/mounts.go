package dockerclient

import (
	"path"

	"github.com/Sirupsen/logrus"
)

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

// AllVolumesmounted checks that all the volumes defined in an image are appropriately mounted in our central persistence directory
func AllVolumesMounted(containerName, persistenceDir, imageName string, volumes map[string]string) bool {

	image, err := InspectImage(imageName)
	if err != nil {
		return false
	}

	for volume, _ := range image.Config.Volumes {
		internalMountPath, ok := volumes[volume]
		if !ok {
			logrus.WithFields(logrus.Fields{
				"container_name": containerName,
				"volume":         volume,
			}).Debug("Volume not mounted")
			return false
		}
		expectedMountPath := path.Join(persistenceDir, containerName, volume)
		if internalMountPath != expectedMountPath {
			logrus.WithFields(logrus.Fields{
				"container_name":      containerName,
				"expected_mount_path": expectedMountPath,
				"internal_mount_path": internalMountPath,
				"volume":              volume,
			}).Debug("Volume not mounted at correct path")
			return false
		}
	}

	return true
}
