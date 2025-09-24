package config

import (
	"os"
	"path/filepath"
	"gopkg.in/yaml.v3"
)

// GlobalSettings holds the global configuration settings for the application.
type GlobalSettings struct {
	ArchivalPlansDir string `yaml:"archivalPlansDir"`
	WorkingDir       string `yaml:"workingDir"`
	DbCredentialsPath string `yaml:"dbCredentialsPath"`
}

// Open settings.yaml file and parse into GlobalSettings struct
func LoadGlobalSettings() (*GlobalSettings, error) {
	settingsFilePath, err := filepath.Abs("./settings.yaml")
	if err != nil {
		return nil, err
	}
	file, err := os.ReadFile(settingsFilePath)
	if err != nil {
		return nil, err
	}

	var settings GlobalSettings
	err = yaml.Unmarshal(file, &settings)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}
