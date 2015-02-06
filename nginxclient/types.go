package nginxclient

import (
	"github.com/rehabstudio/oneill/config"
	"github.com/rehabstudio/oneill/containerdefs"
)

// cfWriter defines a function type that is used for writing nginx
// configuration or htpasswd files to disk
type cfWriter func(*config.Configuration, *containerdefs.RunningContainerDefinition) (bool, error)

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
