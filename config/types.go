package config

type Configuration struct {
	LogFormat            string                         `yaml:"log_format,omitempty"`
	LogLevel             string                         `yaml:"log_level,omitempty"`
	DefinitionsURI       string                         `yaml:"definitions_uri,omitempty"`
	DockerApiEndpoint    string                         `yaml:"docker_api_endpoint,omitempty"`
	PersistenceDirectory string                         `yaml:"persistence_directory,omitempty"`
	RegistryCredentials  map[string]RegistryCredentials `yaml:"registry_credentials"`
}

type RegistryCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
