package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

type FilesConfig struct {
	BucketName      string
	CredentialsFile string
	Timeout         time.Duration
}

var (
	storageConfig FilesConfig
	configInit    sync.Once
	isInitialized bool
)

func InitializeFiles(cfg FilesConfig) error {
	configInit.Do(func() {
		if cfg.BucketName == "" {
			configError = fmt.Errorf("bucket name cannot be empty")
			return
		}
		if cfg.CredentialsFile == "" {
			configError = fmt.Errorf("credentials file path cannot be empty")
			return
		}

		if cfg.Timeout == 0 {
			cfg.Timeout = 10 * time.Second
		}

		storageConfig = cfg
		isInitialized = true
		log.Println("Storage initialized successfully")
	})
	return configError
}

func InitializeStorageClient() (*storage.Client, error) {
	if !isInitialized {
		return nil, fmt.Errorf("storage not initialized. Call Initialize() first")
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(storageConfig.CredentialsFile))
	if err != nil {
		return nil, fmt.Errorf("error initializing storage client: %v", err)
	}
	return client, nil
}

func UploadFile(file multipart.File, fileName string) (string, string, error) {
	if !isInitialized {
		return "", "", fmt.Errorf("storage not initialized. Call Initialize() first")
	}

	id := uuid.New()
	newFileName := id.String() + fileName

	client, err := InitializeStorageClient()
	if err != nil {
		return "", "", err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), storageConfig.Timeout)
	defer cancel()

	bucket := client.Bucket(storageConfig.BucketName)
	object := bucket.Object(newFileName)
	writer := object.NewWriter(ctx)

	writer.ObjectAttrs.Metadata = map[string]string{"firebaseStorageDownloadTokens": id.String()}
	defer writer.Close()

	if _, err := io.Copy(writer, file); err != nil {
		return "", "", fmt.Errorf("failed to upload file: %v", err)
	}

	if err := writer.Close(); err != nil {
		return "", "", fmt.Errorf("failed to finalize upload: %v", err)
	}

	if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", "", fmt.Errorf("failed to set ACL: %v", err)
	}

	fileURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media&token=%s",
		storageConfig.BucketName, newFileName, id.String())

	return fileURL, newFileName, nil
}

func UploadFileWithCustomName(file multipart.File, fileName string) (string, error) {
	if !isInitialized {
		return "", fmt.Errorf("storage not initialized. Call Initialize() first")
	}

	client, err := InitializeStorageClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	id := uuid.New()
	ctx, cancel := context.WithTimeout(context.Background(), storageConfig.Timeout)
	defer cancel()

	bucket := client.Bucket(storageConfig.BucketName)
	object := bucket.Object(fileName)
	writer := object.NewWriter(ctx)

	writer.ObjectAttrs.Metadata = map[string]string{"firebaseStorageDownloadTokens": id.String()}
	defer writer.Close()

	if _, err := io.Copy(writer, file); err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to finalize upload: %v", err)
	}

	if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("failed to set ACL: %v", err)
	}

	fileURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media&token=%s",
		storageConfig.BucketName, fileName, id.String())

	return fileURL, nil
}

func DeleteFile(fileName string) (string, error) {
	if !isInitialized {
		return "", fmt.Errorf("storage not initialized. Call Initialize() first")
	}

	client, err := InitializeStorageClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), storageConfig.Timeout)
	defer cancel()

	bucket := client.Bucket(storageConfig.BucketName)
	object := bucket.Object(fileName)

	if err := object.Delete(ctx); err != nil {
		return "", fmt.Errorf("failed to delete file: %v", err)
	}

	return "File deleted successfully", nil
}

func DownloadFile(fileName string) (string, error) {
	if !isInitialized {
		return "", fmt.Errorf("storage not initialized. Call Initialize() first")
	}

	client, err := InitializeStorageClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), storageConfig.Timeout)
	defer cancel()

	bucket := client.Bucket(storageConfig.BucketName)
	object := bucket.Object(fileName)

	reader, err := object.NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %v", err)
	}

	return string(content), nil
}

func DownloadFileBytes(fileName string) ([]byte, error) {
	if !isInitialized {
		return nil, fmt.Errorf("storage not initialized. Call Initialize() first")
	}

	client, err := InitializeStorageClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), storageConfig.Timeout)
	defer cancel()

	bucket := client.Bucket(storageConfig.BucketName)
	object := bucket.Object(fileName)

	reader, err := object.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}

	return content, nil
}

func FileExists(fileName string) (bool, error) {
	if !isInitialized {
		return false, fmt.Errorf("storage not initialized. Call Initialize() first")
	}

	client, err := InitializeStorageClient()
	if err != nil {
		return false, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), storageConfig.Timeout)
	defer cancel()

	bucket := client.Bucket(storageConfig.BucketName)
	object := bucket.Object(fileName)

	_, err = object.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %v", err)
	}

	return true, nil
}

func GetFileMetadata(fileName string) (map[string]string, error) {
	if !isInitialized {
		return nil, fmt.Errorf("storage not initialized. Call Initialize() first")
	}

	client, err := InitializeStorageClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), storageConfig.Timeout)
	defer cancel()

	bucket := client.Bucket(storageConfig.BucketName)
	object := bucket.Object(fileName)

	attrs, err := object.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %v", err)
	}

	return attrs.Metadata, nil
}
