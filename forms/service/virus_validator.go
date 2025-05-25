package service

import "mime/multipart"

type NullVirusScanner struct{}

func (NullVirusScanner) ScanFile(file multipart.File) (bool, error) {
	return file != nil, nil
}

func NewNullVirusScanner() *NullVirusScanner {
	return &NullVirusScanner{}
}
