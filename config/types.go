package config

type ConfigData struct {
	LogLevel             int                            `yaml:"log_level"`
	DefinitionsDirectory string                         `yaml:"definitions_directory"`
	NginxConfigDirectory string                         `yaml:"nginx_config_directory"`
	ServingDomain        string                         `yaml:"serving_domain"`
	RegistryCredentials  map[string]RegistryCredentials `yaml:"registry_credentials"`
}

type RegistryCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
