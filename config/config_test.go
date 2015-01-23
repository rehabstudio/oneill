package config

import (
	"testing"
)

func TestErrorWhenConfigNotFound(t *testing.T) {

	nonExistantPath := "/this/path/doesnt/exist"
	_, err := loadConfig(nonExistantPath)
	if err == nil {
		t.Errorf("Expected error to be returned when attempting to load non-existant config file.")
	}
}
