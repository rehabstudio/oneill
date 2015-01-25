package config

import (
	"flag"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// loadConfigFileFromDisk reads data from the specified yaml file and
// unmarshalls it into a config struct
func loadConfigFromDisk(path string) (*Configuration, error) {

	// read file from disk
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return &Configuration{}, err
	}

	return loadConfigFromBytes(data)
}

// loadConfigFromBytes takes a slice of bytes and unmarshalls the data into a
// ConfigData struct so it can be used for configuration purposes
func loadConfigFromBytes(data []byte) (cd *Configuration, err error) {

	// unmarshall yaml data to struct
	err = yaml.Unmarshal(data, &cd)
	if err != nil {
		return cd, err
	}

	return cd, nil
}

// loadDefaultConfig initialises a config struct and populates it with default values
func loadDefaultConfig(config *Configuration) *Configuration {

	if isZero(config.DefinitionsURI) {
		config.DefinitionsURI = defaultDefinitionsURI
	}
	if isZero(config.DockerApiEndpoint) {
		config.DockerApiEndpoint = defaultDockerApiEndpoint
	}
	if isZero(config.NginxConfigDirectory) {
		config.NginxConfigDirectory = defaultNginxConfigDirectory
	}
	if isZero(config.NginxHtpasswdDirectory) {
		config.NginxConfigDirectory = defaultNginxHtpasswdDirectory
	}
	if isZero(config.NginxSSLDisabled) {
		config.NginxSSLDisabled = defaultNginxSSLDisabled
	}
	if isZero(config.NginxSSLCertPath) {
		config.NginxSSLCertPath = defaultNginxSSLCertPath
	}
	if isZero(config.NginxSSLKeyPath) {
		config.NginxSSLKeyPath = defaultNginxSSLKeyPath
	}
	if isZero(config.ServingDomain) {
		config.ServingDomain = defaultServingDomain
	}
	if isZero(config.LogLevel) {
		config.LogLevel = defaultLogLevel
	}

	return config
}

// loadConfig initialises a default config then overrides the default values
// with values from the specified configuration file
func loadConfig(configFilePath string) (*Configuration, error) {

	// read configuration file from disk and unmarshall
	config, err := loadConfigFromDisk(configFilePath)

	// load default configuration
	config = loadDefaultConfig(config)

	return config, err
}

// initialises global configuration by first loading default values and then
// overriding with values from a config file
func LoadConfig() (*Configuration, error) {

	// parse config file location from command line flag
	configFile := flag.String("config", "/etc/oneill/config.yaml", "location of the oneill config file")
	flag.Parse()

	// load config from disk into global struct
	config, err := loadConfig(*configFile)
	if err != nil {
		return config, err
	}
	return config, nil

}
