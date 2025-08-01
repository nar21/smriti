package objectStorage

import (
    "fmt"
)

type S3Driver struct {}

func (s *S3Driver) Upload (filePath string) error {
    fmt.Println("Uploading using S3: ", filePath)
    return nil
}