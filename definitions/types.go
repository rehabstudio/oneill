package definitions

type ContainerDefinition struct {
	Subdomain string   `yaml:"subdomain"`
	Image     string   `yaml:"image"`
	Tag       string   `yaml:"tag"`
	Env       []string `yaml:"env"`
}

type DefinitionLoader interface {
	LoadContainerDefinitions() ([]*ContainerDefinition, error)
	ValidateURI() error
}
