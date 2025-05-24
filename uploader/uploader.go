package uploader

import (
	"io"
)

const (
	AWSProvider          = "aws"
	DigitalOceanProvider = "digitalocean"
	LocalProvider        = "local"
)

// FileInfo is a struct that contains information about a file
type FileInfo struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
	Bucket   string `json:"bucket"`
}

// FileUploader is an interface that defines the methods that must be implemented by a file uploader
type FileUploader interface {
	Upload(file io.Reader, bucket string, path string) error
	Info(bucket string, path string) (*FileInfo, error)
	Delete(bucket string, path string) error
	GetTempURL(bucket string, path string) (string, error)
}
