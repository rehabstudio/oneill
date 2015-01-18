package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"text/template"

	"github.com/rehabstudio/oneill/logger"
)

const (
	nginxTemplate string = `
        upstream {{.Subdomain}} {
          server localhost:{{.Port}};
        }

        map $http_upgrade $connection_upgrade {
            default upgrade;
            ''      close;
        }

        server {
          listen *:80;
          server_name {{.Subdomain}}.{{.Domain}};
          return 301 https://$server_name$request_uri;
        }

        server {
          listen 443;
          server_name {{.Subdomain}}.{{.Domain}};

          ssl on;
          ssl_certificate /etc/ssl/certs/labs-server.crt;
          ssl_certificate_key /etc/ssl/private/labs-server.pem;
          ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
          ssl_session_timeout 5m;
          ssl_session_cache shared:SSL:5m;

          client_max_body_size 0; # disable any limits to avoid HTTP 413 for large image uploads

          # required to avoid HTTP 411: see Issue #1486 (https://github.com/docker/docker/issues/1486)
          chunked_transfer_encoding on;

          location / {
            proxy_pass                       http://{{.Subdomain}};
            proxy_set_header  Host           $http_host;   # required for docker client's sake
            proxy_set_header  X-Real-IP      $remote_addr; # pass on real client's IP
            proxy_read_timeout               900;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
          }

        }
`
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
	return runCmd.Run()
}

// templateContext is a simple struct used to contain context
// data for use when rendering templates
type templateContext struct {
	Subdomain string
	Domain    string
	Port      int64
}

// WriteConfig generates an nginx config file to allow reverse proxying into running
// containers. The template is loaded, populated with data and then written to disk.
func WriteConfig(directory string, domain string, subdomain string, port int64) error {
	logger.L.Debug(fmt.Sprintf("Writing nginx configuration for %s.%s", subdomain, domain))

	tmpl, err := template.New("nginx-config").Parse(nginxTemplate)
	if err != nil {
		logger.L.Error(fmt.Sprintf("Unable to load nginx config template: %s", subdomain))
		return err
	}

	// build template context and render the template to `b`
	var b bytes.Buffer
	context := templateContext{Subdomain: subdomain, Domain: domain, Port: port}
	err = tmpl.Execute(&b, context)
	if err != nil {
		logger.L.Error(fmt.Sprintf("Unable to execute nginx config template: %s", subdomain))
		return err
	}

	// write rendered template to disk
	err = ioutil.WriteFile(path.Join(directory, fmt.Sprintf("%s.conf", subdomain)), b.Bytes(), 0644)
	if err != nil {
		logger.L.Error(fmt.Sprintf("Unable to write nginx config template: %s", subdomain))
		return err
	}

	return nil
}
