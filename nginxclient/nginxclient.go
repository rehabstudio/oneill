package nginxclient

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"

	"github.com/rehabstudio/oneill/config"
	"github.com/rehabstudio/oneill/containerdefs"
)

// ConfigureAndReload writes configuration and htpasswd files for all running
// containers before reloading nginx's configuration. This is a destructive
// operation as some files may be overwritten and others removed, it is
// important that oneill is configured correctly and has very sensible
// defaults to account for any silliness here.
func ConfigureAndReload(c *config.Configuration, rcs []*containerdefs.RunningContainerDefinition) error {

	// write nginx configuration file for each running container, overwriting
	// old files if necessary. We return a list of filenames written which
	// we'll use later.
	err := writeNewFiles(c.NginxConfigDirectory, writeNewConfigFile, c, rcs)
	if err != nil {
		return err
	}

	// write htpasswd file for each container that requires it, overwriting
	// old files if necessary. We return a list of filenames written which
	// we'll use later.
	err = writeNewFiles(c.NginxHtpasswdDirectory, writeNewHtpasswdFile, c, rcs)
	if err != nil {
		return err
	}

	// remove redundant configuration files from the config directory. Note
	// that this won't immediately disable the old sites as nginx keeps its
	// configuration in memory and only reloads it when asked.
	err = removeOldFiles(c.NginxConfigDirectory, rcs)
	if err != nil {
		return err
	}

	// remove redundant htpasswd files from the htpasswd directory.
	err = removeOldFiles(c.NginxHtpasswdDirectory, rcs)
	if err != nil {
		return err
	}

	// reload nginx's configuration by sending a HUP signal to the master
	// process, this performs a hot-reload without any downtime
	return reloadNginxConfiguration()
}

// reloadNginxConfiguration issues a `service nginx reload` which causes nginx
// to re-read all of it's configuration files and perform a hot reload. Since
// only root can call this command we use sudo with the `-n` flag, this means
// the the user running oneill is required to have the permission to run this
// command using sudo *without* a password.
func reloadNginxConfiguration() error {

	runCmd := exec.Command("sudo", "-n", "service", "nginx", "reload")
	output, err := runCmd.CombinedOutput()
	if err != nil {
		return err
	}

	// for some reason when `service nginx reload` fails on ubuntu it returns
	// with an exit code of 0. This means we need to parse the commands output
	// to check if it actually failed or not.
	if strings.Contains(string(output[:]), "fail") {
		return errors.New("Failed to reload nginx")
	}

	logrus.Debug("Reloaded nginx configuration")
	return nil
}

// removeIfRedundant checks the given file against a list of currently running
// containers, removing it if a match is not found.
func removeIfRedundant(directory string, f os.FileInfo, rcs []*containerdefs.RunningContainerDefinition) error {

	// if filename matches the name of a currently running container then we
	// just return immediately and skip it.
	for _, rc := range rcs {
		if f.Name() == rc.Name {
			if !rc.ContainerDefinition.NginxDisabled {
				return nil
			}
		}
	}

	filePath := path.Join(directory, f.Name())
	logrus.WithFields(logrus.Fields{"path": filePath}).Info("Removing file")
	return os.Remove(filePath)
}

// removeOldFiles scans a local directory, removing any files where the
// filename does not match the name of a currently running container.
func removeOldFiles(directory string, rcs []*containerdefs.RunningContainerDefinition) error {

	// scan the configured directory, erroring if we don't have permission, it
	// doesn't exist, etc.
	dirContents, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	// loop over all files in the directory checking each one against our
	// currently running list of containers. If the file doesn't match a
	// running container then we delete it.
	for _, f := range dirContents {
		err = removeIfRedundant(directory, f, rcs)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeIfChanged writes the given `content` to disk at `path` if the file
// does not already exist. If the file does already exist then it will only be
// written to if the content is different from what's on disk.
func writeIfChanged(path string, content []byte) error {

	var fileExists bool
	var contentChanged bool

	if _, err := os.Stat(path); err == nil {
		fileExists = true

		readContent, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if !bytes.Equal(content, readContent) {
			contentChanged = true
		}
	}

	if !fileExists || contentChanged {
		logrus.WithFields(logrus.Fields{"path": path}).Info("Writing file")
		return ioutil.WriteFile(path, content, 0644)
	}

	return nil
}

// writeNewConfigFile writes a new nginx configuration file to disk for the
// given container definition. A simple template file is used which is
// compiled into the binary at build time. A new file will only be written if
// the file either doesn't exist or its contents have changed.
func writeNewConfigFile(c *config.Configuration, rc *containerdefs.RunningContainerDefinition) error {

	// check if this config file needs to reference a htpasswd file or not
	hasHtpasswd := len(rc.ContainerDefinition.Htpasswd) > 0
	htpasswdFile := path.Join(c.NginxHtpasswdDirectory, rc.Name)

	// load configuration file template so we can render it
	nginxTemplateBytes, err := Asset("templates/reverse_proxy.tmpl")
	if err != nil {
		return err
	}
	nginxTemplate, err := template.New("nginx").Parse(string(nginxTemplateBytes[:]))
	if err != nil {
		return err
	}

	// build template context and render the template to `b`
	var b bytes.Buffer
	context := templateContext{
		Subdomain:    rc.ContainerDefinition.Subdomain,
		HasHtpasswd:  hasHtpasswd,
		HtpasswdFile: htpasswdFile,
		SSLDisabled:  c.NginxSSLDisabled,
		SSLCertPath:  c.NginxSSLCertPath,
		SSLKeyPath:   c.NginxSSLKeyPath,
		Domain:       c.ServingDomain,
		Port:         rc.Port,
	}
	if nginxTemplate.Execute(&b, context) != nil {
		return err
	}

	// write rendered template to disk
	configFilePath := path.Join(c.NginxConfigDirectory, rc.Name)
	return writeIfChanged(configFilePath, b.Bytes())
}

// writeNewFiles writes a file to disk for each running container using the
// passed in function. writeNewFiles first ensures that the directory into
// which the files will be written has been created.
func writeNewFiles(d string, f cfWriter, c *config.Configuration, rcs []*containerdefs.RunningContainerDefinition) error {

	// create directory to store config/htpasswd files
	err := os.MkdirAll(d, 0755)
	if err != nil {
		return err
	}

	// loop over and write a configuration file for every running container
	for _, rc := range rcs {
		// if this container definition has explicitly disabled nginx support
		// then we just skip it
		if rc.ContainerDefinition.NginxDisabled {
			continue
		}
		// call the passed in cfWriter function on each container
		err = f(c, rc)
		if err != nil {
			return err
		}
	}
	return nil
}

// writeNewHtpasswdFile writes a htpasswd file to disk if required. A new file
// will only be written if the file either doesn't exist or its contents have
// changed.
func writeNewHtpasswdFile(c *config.Configuration, rc *containerdefs.RunningContainerDefinition) error {

	// check if we need to write a htpasswd file or not
	if len(rc.ContainerDefinition.Htpasswd) == 0 {
		return nil
	}

	// write htpasswd file to disk
	htpasswdFilePath := path.Join(c.NginxHtpasswdDirectory, rc.Name)
	fileContent := []byte(strings.Join(rc.ContainerDefinition.Htpasswd, "\n"))
	return writeIfChanged(htpasswdFilePath, fileContent)
}
