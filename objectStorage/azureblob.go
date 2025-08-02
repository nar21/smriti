package objectStorage

import (
	"fmt"
)

type AzureBlobDriver struct{}

func (s *AzureBlobDriver) Upload(filePath string, remoteFilePath string) error {
	fmt.Println("Uploading using Azure Blob: ", filePath)
	return nil
}
