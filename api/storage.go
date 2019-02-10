package api

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// StorageConfig represents configuration for creating a new Storage
type StorageConfig struct {
	AccessKey    string
	AccessSecret string
	Region       string
	Bucket       string
}

// Storage represents an interface pointing to an external file host (AWS S3, in this case)
type Storage struct {
	AWS     *s3.S3
	session *session.Session
	config  StorageConfig
}

// NewStorage creates a new Storage from a passed in AWS config
func NewStorage(config StorageConfig) (*Storage, error) {
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(config.AccessKey, config.AccessSecret, ""),
	})
	if err != nil {
		return nil, err
	}
	return &Storage{
		AWS:     s3.New(session),
		session: session,
		config:  config,
	}, nil
}

// Upload uploads a file to AWS S3 through the Storage interface, returning the resource URL
func (s *Storage) Upload(file io.Reader, key string) (string, error) {
	uploader := s3manager.NewUploader(s.session)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.config.Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://s3.%s.amazonaws.com/%s/%s", s.config.Region, s.config.Bucket, key), nil
}

// Delete removes a file from AWS S3 through the Storage interface
func (s *Storage) Delete(key string) error {
	_, err := s.AWS.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	err = s.AWS.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}
