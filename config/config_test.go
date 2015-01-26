package config

import (
	"reflect"
	"testing"
)

func TestErrorWhenConfigNotFound(t *testing.T) {

	nonExistantPath := "/this/path/doesnt/exist"
	_, err := loadConfig(nonExistantPath)
	if err == nil {
		t.Errorf("Expected error to be returned when attempting to load non-existant config file.")
	}
}

func mergeAndCompare(t *testing.T, note string, input, expected *Configuration) {
	config := mergeConfigs(loadDefaultConfig(), input)
	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Actual Config != Expected config (%s).", note)
	}
}

func TestMergeExampleConfig1(t *testing.T) {

	input := &Configuration{NginxConfigDirectory: "/etc/nginx/conf.d"}
	expected := loadDefaultConfig()
	expected.NginxConfigDirectory = "/etc/nginx/conf.d"
	mergeAndCompare(t, "Test 1", input, expected)
}

func TestMergeExampleConfig2(t *testing.T) {

	input := &Configuration{NginxSSLDisabled: true}
	expected := loadDefaultConfig()
	expected.NginxSSLDisabled = true
	mergeAndCompare(t, "Test 2", input, expected)
}

func TestMergeExampleConfig3(t *testing.T) {

	input := &Configuration{
		NginxConfigDirectory: "/etc/nginx/conf.d",
		DockerApiEndpoint:    "",
	}
	expected := loadDefaultConfig()
	expected.NginxConfigDirectory = "/etc/nginx/conf.d"
	mergeAndCompare(t, "Test 3", input, expected)
}

func TestMergeExampleConfig4(t *testing.T) {

	input := &Configuration{
		NginxConfigDirectory: "/etc/nginx/conf.d",
		NginxSSLCertPath:     "/etc/ssl/test.crt",
		NginxSSLKeyPath:      "/etc/ssl/test.pem",
	}
	expected := loadDefaultConfig()
	expected.NginxConfigDirectory = "/etc/nginx/conf.d"
	expected.NginxSSLCertPath = "/etc/ssl/test.crt"
	expected.NginxSSLKeyPath = "/etc/ssl/test.pem"
	mergeAndCompare(t, "Test 4", input, expected)
}

func TestMergeExampleConfig5(t *testing.T) {

	input := &Configuration{}
	expected := loadDefaultConfig()
	mergeAndCompare(t, "Test 5", input, expected)
}
