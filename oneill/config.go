package oneill

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

var Config *ConfigData

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

func loadConfigFile(path string) (*ConfigData, error) {
	cd := ConfigData{}
	// read file from disk
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return &cd, err
	}
	// unmarshall yaml data to struct
	err = yaml.Unmarshal(data, &cd)
	if err != nil {
		return &cd, err
	}
	return &cd, nil
}

func loadDefaults() {
	if Config.DefinitionsDirectory == "" {
		Config.DefinitionsDirectory = "/etc/oneill/definitions"
	}
	if Config.NginxConfigDirectory == "" {
		Config.NginxConfigDirectory = "/etc/nginx/sites-enabled"
	}
	if Config.ServingDomain == "" {
		Config.ServingDomain = "example.com"
	}
	Config.RegistryCredentials[""] = RegistryCredentials{}
}

func InitConfig() {

	configFile := flag.String("config", "/etc/oneill/config.yaml", "location of the oneill config file")
	flag.Parse()

	config, err := loadConfigFile(*configFile)
	if err != nil {
		// we use baseLogger directly because until config is parsed our
		// logging setup won't work correctly
		baseLogger("FATAL", fmt.Sprintf("Error loading configuration file: %s", err))
		os.Exit(1)
	}
	Config = config
	loadDefaults()

}
