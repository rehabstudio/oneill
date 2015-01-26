package config

type Configuration struct {
	LogLevel               string                         `yaml:"log_level,omitempty"`
	DefinitionsURI         string                         `yaml:"definitions_uri,omitempty"`
	DockerApiEndpoint      string                         `yaml:"docker_api_endpoint,omitempty"`
	NginxConfigDirectory   string                         `yaml:"nginx_config_directory,omitempty"`
	NginxHtpasswdDirectory string                         `yaml:"nginx_htpasswd_directory,omitempty"`
	ServingDomain          string                         `yaml:"serving_domain,omitempty"`
	RegistryCredentials    map[string]RegistryCredentials `yaml:"registry_credentials"`
	NginxSSLDisabled       bool                           `yaml:"nginx_ssl_disabled"`
	NginxSSLCertPath       string                         `yaml:"nginx_ssl_cert"`
	NginxSSLKeyPath        string                         `yaml:"nginx_ssl_key"`
}

type RegistryCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
