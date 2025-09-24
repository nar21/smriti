package objectStorage

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"smriti/logging"
)

type S3Driver struct {
	Bucket string
	Region string
	Logger *logging.JobLogger
	Client *s3.Client
}

// NewS3Driver initializes the S3Driver with AWS config and bucket name.
func NewS3Driver(bucket string, region string, logger *logging.JobLogger) (*S3Driver, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	return &S3Driver{
		Bucket: bucket,
		Region: region,
		Logger: logger,
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

	s.Logger.Log(
		"Uploaded",
		filePath,
		"to",
		s.Bucket,
		"://",
		remoteFilePath,
	)
	return nil
}
