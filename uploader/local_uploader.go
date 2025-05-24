package uploader

import (
	"io"
	"mime"
	"os"
	"path/filepath"
)

// LocalUploader is a struct that contains information about a file
// This uploader is used for local development only
type LocalUploader struct {
	UploadDirectory string
}

// NewLocalUploader creates a new local uploader
func NewLocalUploader(uploadDirectory string) *LocalUploader {
	return &LocalUploader{UploadDirectory: uploadDirectory}
}

func (u *LocalUploader) Upload(file io.Reader, _ string, path string) error {
	outputPath := filepath.Join(u.UploadDirectory, path)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		return err
	}
	return nil
}

func (u *LocalUploader) Info(_ string, path string) (*FileInfo, error) {
	filePath := filepath.Join(u.UploadDirectory, path)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	mimeType := mime.TypeByExtension(filepath.Ext(path))
	return &FileInfo{URL: path, Bucket: u.UploadDirectory, MimeType: mimeType, Size: fileInfo.Size()}, nil
}

func (u *LocalUploader) Delete(bucket string, path string) error {
	filePath := filepath.Join(bucket, path)
	return os.Remove(filePath)
}

func (u *LocalUploader) GetTempURL(bucket string, path string) (string, error) {
	return "", nil
}
