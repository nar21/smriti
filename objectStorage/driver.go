package objectStorage

type ObjectStorageDriver interface {
    Upload(filePath string, remoteFilePath string) error
}