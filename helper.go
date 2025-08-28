package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"time"
)

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
