package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/rehabstudio/oneill/logger"
)

// global ConfigData struct used to hold configuration data used by the app.
// *must* be initialised via InitConfig() before use.
var Config *ConfigData

// loadConfigFileFromDisk reads data from the specified yaml file and
// unmarshalls it into a config struct
func loadConfigFromDisk(path string, config *ConfigData) (*ConfigData, error) {

	// read file from disk
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return &ConfigData{}, err
	}

	return loadConfigFromBytes(data)
}

// loadConfigFromBytes takes a slice of bytes and unmarshalls the data into a
// ConfigData struct so it can be used for configuration purposes
func loadConfigFromBytes(data []byte) (cd *ConfigData, err error) {

	// unmarshall yaml data to struct
	err = yaml.Unmarshal(data, &cd)
	if err != nil {
		return cd, err
	}

	return cd, nil
}

// loadDefaultConfig initialises a config struct and populates it with default values
func loadDefaultConfig() *ConfigData {

	config := ConfigData{}
	config.DefinitionsDirectory = defaultDefinitionsDirectory
	config.NginxConfigDirectory = defaultNginxConfigDirectory
	config.ServingDomain = defaultServingDomain
	config.LogLevel = defaultLogLevel

	return &config
}

// loadConfig initialises a default config then overrides the default values
// with values from the specified configuration file
func loadConfig(configFilePath string) (*ConfigData, error) {

	// load default configuration
	config := loadDefaultConfig()
	// read configuration file from disk and unmarshall (overwriting any default settings)
	config, err := loadConfigFromDisk(configFilePath, config)

	return config, err
}

// initialises global configuration by first loading default values and then
// overriding with values from a config file
func InitConfig() {

	// parse config file location from command line flag
	configFile := flag.String("config", "/etc/oneill/config.yaml", "location of the oneill config file")
	flag.Parse()

	// load config from disk into global struct
	config, err := loadConfig(*configFile)
	if err != nil {
		logger.LogFatal(fmt.Sprintf("Error loading configuration file: %s", err))
		os.Exit(1)
	}
	Config = config

}
