package containerdefs

type ContainerDefinition struct {
	Subdomain     string   `yaml:"subdomain"`
	Image         string   `yaml:"image"`
	Tag           string   `yaml:"tag"`
	NginxDisabled bool     `yaml:"nginx_disabled"`
	Env           []string `yaml:"env"`
	Htpasswd      []string `yaml:"htpasswd"`
}

type RunningContainerDefinition struct {
	ContainerDefinition *ContainerDefinition
	Name                string
	Port                int64
}

type ContainerProcessor interface {
	GetExistingContainer(*ContainerDefinition) (*RunningContainerDefinition, error)
	PullImage(*ContainerDefinition) error
	ValidateImage(*ContainerDefinition) error
	StartContainer(*ContainerDefinition) (*RunningContainerDefinition, error)
}

type DefinitionLoader interface {
	LoadContainerDefinitions() ([]*ContainerDefinition, error)
	ValidateURI() error
}
