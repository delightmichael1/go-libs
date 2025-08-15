package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func init() {
	if os.Getenv("ENVIRONMENT") == "" || os.Getenv("ENVIRONMENT") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}
}

func InitializeStorageClient() (*storage.Client, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("storage.config.json"))
	if err != nil {
		return nil, fmt.Errorf("error initializing storage client: %v", err)
	}
	return client, nil
}

func UploadFile(file multipart.File, fileName string) (string, string, error) {
	id := uuid.New()
	newFileName := id.String() + fileName
	client, err := InitializeStorageClient()
	if err != nil {
		return "", "", err
	}
	defer client.Close()

	bucketName := os.Getenv("STORAGEBUCKET")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	bucket := client.Bucket(bucketName)
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

	// fileURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media&token=%s", bucketName, fileName, id.String())
	fileURL := "https://firebasestorage.googleapis.com/v0/b/" + bucketName + "/o/" + newFileName + "?alt=media&token=" + id.String()

	return fileURL, newFileName, nil
}

func DeleteFile(fileName string) (string, error) {
	client, err := InitializeStorageClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	bucketName := os.Getenv("STORAGEBUCKET")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	bucket := client.Bucket(bucketName)
	object := bucket.Object(fileName)

	if err := object.Delete(ctx); err != nil {
		return "", fmt.Errorf("failed to delete file: %v", err)
	}

	return "File deleted successfully", nil
}

func DownloadFile(fileName string) (string, error) {
	client, err := InitializeStorageClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	bucketName := os.Getenv("STORAGEBUCKET")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	bucket := client.Bucket(bucketName)
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
