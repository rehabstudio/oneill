package definitions

const (
	defaultTag string = "latest"
)

// loadContainerDefaults fills in any blanks in the definition
func loadContainerDefaults(cd *ContainerDefinition) *ContainerDefinition {

	if cd.Tag == "" {
		cd.Tag = "latest"
	}

	return cd
}
