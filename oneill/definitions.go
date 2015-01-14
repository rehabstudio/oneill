package oneill

type SiteConfig struct {
	Subdomain string `yaml:"subdomain"`
	Container string `yaml:"container"`
	Tag       string `yaml:"tag"`
}
