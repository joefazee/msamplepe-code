package uploader

import (
	"io"
	"mime"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const TempURLTimePeriod = 30 * time.Minute

// S3Uploader is a struct that contains information about a file
type S3Uploader struct {
	Uploader *s3manager.Uploader
	S3Client *s3.S3
}

// NewS3Uploader creates a new s3 uploader
func NewS3Uploader(session *session.Session) *S3Uploader {
	uploader := s3manager.NewUploader(session)
	return &S3Uploader{Uploader: uploader, S3Client: s3.New(session)}
}

// Upload uploads a file to s3
func (u *S3Uploader) Upload(file io.Reader, bucket string, path string) error {
	_, err := u.Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		Body:   file,
	})
	if err != nil {
		return err
	}
	return nil
}

// Info returns information about a file in s3
func (u *S3Uploader) Info(bucket string, path string) (*FileInfo, error) {
	mimeType := mime.TypeByExtension(filepath.Ext(path))
	return &FileInfo{URL: path, Bucket: bucket, MimeType: mimeType, Size: 0}, nil
}

func (u *S3Uploader) Delete(bucket string, path string) error {
	_, err := u.S3Client.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(path)})
	if err != nil {
		return err
	}

	err = u.S3Client.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	})

	return err
}

func (u *S3Uploader) GetTempURL(bucket string, path string) (string, error) {
	req, _ := u.S3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	})

	url, err := req.Presign(TempURLTimePeriod)
	if err != nil {
		return "", err
	}

	return url, nil
}
