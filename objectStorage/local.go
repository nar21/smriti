package objectStorage

import (
	"fmt"
	"os"
	"path/filepath"
)

type LocalFSDriver struct {
	Path string
}

// NewS3Driver initializes the S3Driver with AWS config and bucket name.
func NewLocalFSDriver(path string) (*LocalFSDriver, error) {
	return &LocalFSDriver{
		Path: path,
	}, nil
}

// Upload uploads a file to the configured S3 bucket.
func (s *LocalFSDriver) Upload(src string, destDir string) error {
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

	destDir, err = filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of destination directory: %w", err)
	}

	// Ensure the destination is a directory and it exists
	destFd, err := os.Stat(destDir)
	if err != nil {
		fmt.Println("Destination does not exist: ", destDir)
	}
	if !destFd.IsDir() {
		return fmt.Errorf("destination %s is not a directory", destDir)
	}

	// Copy file to the destination file path
	destFile := filepath.Join(destDir, filepath.Base(src))
	err = os.WriteFile(destFile, srcFile, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Copied %s to %s\n", srcFile, destFile)
	return nil
}
