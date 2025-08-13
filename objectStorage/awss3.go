package objectStorage

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Driver struct {
	Bucket string `yaml:"bucket"`
	Region string `yaml:"region"`
	Client *s3.Client
}

// NewS3Driver initializes the S3Driver with AWS config and bucket name.
func NewS3Driver(bucket string) (*S3Driver, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	return &S3Driver{
		Bucket: bucket,
		Client: client,
	}, nil
}

// Upload uploads a file to the configured S3 bucket.
func (s *S3Driver) Upload(filePath string, remoteFilePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(remoteFilePath),
		Body:   file,
		ACL:    types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return err
	}

	fmt.Printf(
		"Uploaded %s to s3://%s/%s\n",
		filePath,
		s.Bucket,
		remoteFilePath,
	)
	return nil
}
