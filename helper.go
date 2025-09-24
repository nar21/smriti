package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"smriti/config"
	"smriti/parser"
	"time"

	"gopkg.in/yaml.v3"
)

// Function to initialize application directories and files
func initSmritiDirsAndFiles() error {
	// Create settings.yaml with default values and exit
	settingsFilePath, err := filepath.Abs("./settings.yaml")
	if err != nil {
		log.Fatal("Could not get absolute path of settings.yaml")
	}

	// Initialize default values
	initArchivalPlathDir := "./archival-plan"
	initWorkingDir := "./workDir"
	initDBCredentialsPath := "./db-credentials.yaml"

	// Create a default settings.yaml file

	if _, err := os.Stat(settingsFilePath); os.IsNotExist(err) {
		defaultSettings := config.GlobalSettings{
			ArchivalPlansDir:  initArchivalPlathDir,
			WorkingDir:        initWorkingDir,
			DbCredentialsPath: initDBCredentialsPath,
		}

		file, err := os.Create(settingsFilePath)
		if err != nil {
			return fmt.Errorf("could not create settings.yaml file")
		}
		defer file.Close()

		data, err := yaml.Marshal(&defaultSettings)
		if err != nil {
			return fmt.Errorf("could not marshal default settings to YAML")
		}

		_, err = file.Write(data)
		if err != nil {
			return fmt.Errorf("could not write default settings to settings.yaml file")
		}

		fmt.Println("Created default settings.yaml at ", settingsFilePath)
		// Continue execution to create db-credentials.yaml file
	} else {
		return fmt.Errorf("settings.yaml already exists at %s", settingsFilePath)
	}

	// Create a default db-credentials.yaml file
	credFilePath, err := filepath.Abs(initDBCredentialsPath)
	if err != nil {
		return fmt.Errorf("could not get absolute path of db-credentials.yaml")
	}

	if _, err := os.Stat(credFilePath); os.IsNotExist(err) {
		defaultCreds := parser.CredentialFile{
			Databases: map[string]parser.DBCredential{
				"sample-rds": {
					Engine:   "postgresql",
					Host:     "your-db-host",
					DBName:   "your-db-name",
					User:     "your-username",
					Password: "your-password",
					Port:     5432,
				},
			},
		}

		file, err := os.Create(credFilePath)
		if err != nil {
			return fmt.Errorf("could not create db-credentials.yaml file")
		}
		defer file.Close()

		data, err := yaml.Marshal(&defaultCreds)
		if err != nil {
			return fmt.Errorf("could not marshal default DB credentials to YAML")
		}

		_, err = file.Write(data)
		if err != nil {
			return fmt.Errorf("could not write default DB credentials to db-credentials.yaml file")
		}

		fmt.Println("created default db-credentials.yaml at ", credFilePath)
		return nil
	} else {
		return fmt.Errorf("db-credentials.yaml already exists at %s", credFilePath)
	}
	return nil
}

// GenerateDateHexString returns the current date in YYYYMMDD format appended with a random 6-hex-character string.
func GenerateExecutionID() (string, error) {
	// Get current date in YYYYMMDD format
	dateStr := time.Now().Format("2006-01-02")

	// Generate 3 random bytes (6 hex characters)
	randomBytes := make([]byte, 3)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Format bytes as hex string
	randomHex := fmt.Sprintf("%06x", randomBytes)
	return fmt.Sprintf("%s-%s", dateStr, randomHex), nil
}

// CreateDirIfNotExist checks if a directory exists, and creates it if it does not.
func CreateDirIfNotExist(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Directory does not exist, create it (including parents)
		err := os.MkdirAll(dir, 0755) // 0755 is typical permission
		if err != nil {
			return err
		}
		fmt.Println("Directory created:", dir)
	} else {
		fmt.Println("Directory already exists:", dir)
	}
	return nil
}
