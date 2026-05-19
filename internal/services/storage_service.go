package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UploadedFileResult struct {
	Path string
	URL  string
	Disk string
}

type LocalStorageService struct {
	BaseURL string
}

func NewLocalStorageService(baseURL string) *LocalStorageService {
	return &LocalStorageService{
		BaseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (s *LocalStorageService) UploadProductImage(fileHeader *multipart.FileHeader) (*UploadedFileResult, error) {
	if fileHeader == nil {
		return nil, nil
	}

	const maxFileSize = 2 * 1024 * 1024

	if fileHeader.Size > maxFileSize {
		return nil, errors.New("image size must be less than or equal to 2MB")
	}

	extension := strings.ToLower(filepath.Ext(fileHeader.Filename))

	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}

	if !allowedExtensions[extension] {
		return nil, errors.New("image extension must be jpg, jpeg, png, or webp")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	uploadDir := filepath.Join("uploads", "products")

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, err
	}

	randomName, err := generateRandomFileName()
	if err != nil {
		return nil, err
	}

	fileName := "product-" + time.Now().Format("20060102150405") + "-" + randomName + extension
	savePath := filepath.Join(uploadDir, fileName)

	destination, err := os.Create(savePath)
	if err != nil {
		return nil, err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, file); err != nil {
		return nil, err
	}

	normalizedPath := filepath.ToSlash(savePath)
	fileURL := s.BaseURL + "/" + normalizedPath

	return &UploadedFileResult{
		Path: normalizedPath,
		URL:  fileURL,
		Disk: "local",
	}, nil
}

func generateRandomFileName() (string, error) {
	bytes := make([]byte, 8)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
