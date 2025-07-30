package objectStorage

type ObjectStorageDriver interface {
    Upload(filePath string) error
}