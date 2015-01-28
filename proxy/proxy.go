package proxy

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
)

// ClearConfigDirectory ensures an empty directory exists in which to save our configuration files.
func ClearConfigDirectory(directory string) error {

	if err := os.RemoveAll(directory); err != nil {
		return err
	}
	if err := os.Mkdir(directory, 0755); err != nil {
		return err
	}

	return nil
}

// ReloadProxy issues a `service nginx reload` which causes nginx to re-read all
// of it's configuration files and perform a hot reload.
func ReloadServer() error {

	runCmd := exec.Command("service", "nginx", "reload")

	output, err := runCmd.CombinedOutput()
	if err != nil {
		return err
	}

	if strings.Contains(string(output[:]), "fail") {
		return errors.New("Failed to reload nginx")
	}

	return nil
}

// templateContext is a simple struct used to contain context
// data for use when rendering templates
type templateContext struct {
	Subdomain    string
	HtpasswdFile string
	Domain       string
	HasHtpasswd  bool
	SSLDisabled  bool
	SSLCertPath  string
	SSLKeyPath   string
	Port         int64
}

// WriteConfig generates an nginx config file to allow reverse proxying into running
// containers. The template is loaded, populated with data and then written to disk.
func WriteConfig(nginxConfDirectory string, nginxHtpasswdDirectory string, domain string, subdomain string, htpasswd []string, port int64, sslDisabled bool, sslCertPath string, sslKeyPath string) error {

	// create htpasswd file
	var hasHtpasswd bool
	htpasswdFile := path.Join(nginxHtpasswdDirectory, subdomain)
	if len(htpasswd) > 0 {
		c := strings.Join(htpasswd, "\n")
		logrus.WithFields(logrus.Fields{
			"subdomain": subdomain,
			"domain":    domain,
		}).Debug("Writing htpasswd file")
		d := []byte(c)
		err := ioutil.WriteFile(htpasswdFile, d, 0644)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Error("Something went wrong while trying to write the htpasswd file")
			return err
		}
		hasHtpasswd = true
	}

	logrus.WithFields(logrus.Fields{
		"subdomain": subdomain,
		"domain":    domain,
	}).Debug("Writing nginx configuration")

	nginxTemplate, err := Asset("templates/reverse_proxy.tmpl")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unable to load nginx config template")
		return err
	}
	tmpl, err := template.New("nginx-config").Parse(string(nginxTemplate[:]))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unable to load nginx config template")
		return err
	}

	// build template context and render the template to `b`
	var b bytes.Buffer
	context := templateContext{Subdomain: subdomain, HasHtpasswd: hasHtpasswd, HtpasswdFile: htpasswdFile, SSLDisabled: sslDisabled, SSLCertPath: sslCertPath, SSLKeyPath: sslKeyPath, Domain: domain, Port: port}
	err = tmpl.Execute(&b, context)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unable to execute nginx config template")
		return err
	}

	// write rendered template to disk
	err = ioutil.WriteFile(path.Join(nginxConfDirectory, fmt.Sprintf("%s.conf", subdomain)), b.Bytes(), 0644)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unable to write nginx config template")
		return err
	}

	return nil
}
