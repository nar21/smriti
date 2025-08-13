package objectStorage

import (
	"fmt"
	"os"
	"path/filepath"
)

type LocalFSDriver struct {
	Path string `yaml:"path"`
}

// NewS3Driver initializes the S3Driver with AWS config and bucket name.
func NewLocalFSDriver(path string) (*LocalFSDriver, error) {
	return &LocalFSDriver{
		Path: path,
	}, nil
}

// Upload uploads a file to the configured S3 bucket.
func (s *LocalFSDriver) Upload(src string, dest string) error {
	src, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of source file: %w", err)
	}

	// Check if source file exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", src)
	}
	srcFile, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	dest, err = filepath.Abs(fmt.Sprintf("%s/%s", s.Path, dest))
	if err != nil {
		return fmt.Errorf("failed to get absolute path of destination directory: %w", err)
	}

	destDir := filepath.Dir(dest)

	fmt.Println("Destination Directory: ", destDir)
	// Ensure the destination is a directory and it exists
	destFd, err := os.Stat(destDir)

	// If the destination is not a directory, return an error
	if err == nil && !destFd.IsDir() {
		return fmt.Errorf("destination %s is not a directory", destDir)
	}

	// If the directory does not exist, create it
	if err != nil {
		os.MkdirAll(destDir, 0755) // Create the directory if it doesn't exist
		fmt.Println("Created directory:", destDir)
	}

	// Copy file to the destination file path
	err = os.WriteFile(dest, srcFile, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Copied %s to %s\n", src, dest)
	return nil
}
