package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// LoadConfig initialises global configuration by first loading default values
// and then overriding with values from a config file
func LoadConfig(configFilePath string) (*Configuration, error) {

	// load default configuration
	defaultConfig := loadDefaultConfig()

	// read configuration file from disk and unmarshall
	diskConfig, err := loadConfigFromDisk(configFilePath)
	if err != nil {
		return &Configuration{}, err
	}

	// merge user config with default config to get effective configuration
	// for this instance of the application.
	config := mergeConfigs(defaultConfig, diskConfig)

	return config, err
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

// mergeConfigs takes an arbitrary number of configuration structs, iterating
// over them, replacing any "zero" value in the previous struct with a
// non-zero value from the next.
func mergeConfigs(configs ...*Configuration) *Configuration {
	newConfig := &Configuration{}
	for _, config := range configs {
		if !isZero(config.DefinitionsURI) {
			newConfig.DefinitionsURI = config.DefinitionsURI
		}
		if !isZero(config.DockerApiEndpoint) {
			newConfig.DockerApiEndpoint = config.DockerApiEndpoint
		}
		if !isZero(config.LogFormat) {
			newConfig.LogFormat = config.LogFormat
		}
		if !isZero(config.LogLevel) {
			newConfig.LogLevel = config.LogLevel
		}
		if !isZero(config.NginxConfigDirectory) {
			newConfig.NginxConfigDirectory = config.NginxConfigDirectory
		}
		if !isZero(config.NginxHtpasswdDirectory) {
			newConfig.NginxHtpasswdDirectory = config.NginxHtpasswdDirectory
		}
		if !isZero(config.NginxSSLCertPath) {
			newConfig.NginxSSLCertPath = config.NginxSSLCertPath
		}
		if !isZero(config.NginxDisabled) {
			newConfig.NginxDisabled = config.NginxDisabled
		}
		if !isZero(config.NginxSSLDisabled) {
			newConfig.NginxSSLDisabled = config.NginxSSLDisabled
		}
		if !isZero(config.NginxSSLKeyPath) {
			newConfig.NginxSSLKeyPath = config.NginxSSLKeyPath
		}
		if !isZero(config.RegistryCredentials) {
			newConfig.RegistryCredentials = config.RegistryCredentials
		}
		if !isZero(config.ServingDomain) {
			newConfig.ServingDomain = config.ServingDomain
		}
	}

	return newConfig
}
