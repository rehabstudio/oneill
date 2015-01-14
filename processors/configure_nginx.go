package processors

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/fsouza/go-dockerclient"
	"github.com/rehabstudio/oneill/oneill"
)

const (
	nginxTemplate string = `
        upstream {{.Subdomain}} {
          server localhost:{{.Port}};
        }

        server {
          listen *:80;
          server_name {{.Subdomain}}.{{.ServingDomain}};
          return 301 https://$server_name$request_uri;
        }

        server {
          listen 443;
          server_name {{.Subdomain}}.{{.ServingDomain}};

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
          }

        }
`
)

func ensureFreshOutputDir() {
	oneill.LogDebug("Recreating empty configuration directory")
	outDir := oneill.Config.NginxConfigDirectory
	err := os.RemoveAll(outDir)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(outDir, 0755)
	if err != nil {
		panic(err)
	}
}

func reloadNginxConfig() error {
	runCmd := exec.Command("service", "nginx", "reload")
	return runCmd.Run()
}

type templateContext struct {
	Subdomain     string
	ServingDomain string
	Port          int64
}

func writeTemplateToDisk(siteConfig *oneill.SiteConfig, container docker.APIContainers) {
	tmpl, err := template.New("nginx-config").Parse(nginxTemplate)
	if err != nil {
		oneill.LogWarning(fmt.Sprintf("Unable to load nginx config template: %s", siteConfig.Subdomain))
		return
	}

	var b bytes.Buffer
	context := templateContext{
		Subdomain:     siteConfig.Subdomain,
		ServingDomain: oneill.Config.ServingDomain,
		Port:          container.Ports[0].PublicPort,
	}
	err = tmpl.Execute(&b, context)
	if err != nil {
		oneill.LogWarning(fmt.Sprintf("Unable to execute nginx config template: %s", siteConfig.Subdomain))
		return
	}
	outDir := oneill.Config.NginxConfigDirectory
	outFile := path.Join(outDir, fmt.Sprintf("%s.conf", siteConfig.Subdomain))
	err = ioutil.WriteFile(outFile, b.Bytes(), 0644)
	if err != nil {
		oneill.LogWarning(fmt.Sprintf("Unable to write nginx config template: %s", siteConfig.Subdomain))
		return
	}
}

func ConfigureNginx(siteConfigs []*oneill.SiteConfig) []*oneill.SiteConfig {
	oneill.LogInfo("## Configuring Nginx")

	ensureFreshOutputDir()

	for _, sc := range siteConfigs {
		for _, container := range oneill.ListContainers() {
			containerName := strings.TrimPrefix(container.Names[0], "/")
			if containerName == sc.Subdomain {
				writeTemplateToDisk(sc, container)
				oneill.LogDebug(fmt.Sprintf("Configured nginx proxy for container: %s", sc.Subdomain))
				break
			}
		}
	}
	if err := reloadNginxConfig(); err != nil {
		oneill.LogWarning("Unable to reload nginx configuration")
		return siteConfigs
	}
	oneill.LogDebug("Reloaded nginx configuration")
	return siteConfigs
}
