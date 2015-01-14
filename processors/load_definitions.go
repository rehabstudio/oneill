package processors

import (
	"io/ioutil"
	"path"

	"github.com/rehabstudio/oneill/oneill"
	"gopkg.in/yaml.v2"
)

// loadConfig loads a siteconfig.yaml file from disk and unmarshalls it to a
// SiteConfig struct
func loadConfig(path string) (*oneill.SiteConfig, error) {

	sc := oneill.SiteConfig{}

	// read file from disk
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return &sc, err
	}

	// unmarshall yaml data to struct
	err = yaml.Unmarshal(data, &sc)
	if err != nil {
		return &sc, err
	}

	// set tag to "latest" if not set
	if sc.Tag == "" {
		sc.Tag = "latest"
	}

	return &sc, nil

}

func LoadSiteDefinitions(_ []*oneill.SiteConfig) (siteConfigs []*oneill.SiteConfig) {
	dirContents, err := ioutil.ReadDir(oneill.Config.DefinitionsDirectory)
	if err != nil {
		panic(err)
	}
	for _, f := range dirContents {
		if f.IsDir() {
			sc, err := loadConfig(path.Join(oneill.Config.DefinitionsDirectory, f.Name(), "siteconfig.yaml"))
			if err != nil {
				continue
			}
			siteConfigs = append(siteConfigs, sc)
		}
	}
	return siteConfigs
}
