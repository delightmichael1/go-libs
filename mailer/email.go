package mailer

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/gomail.v2"
)

type Config struct {
	SMTPHost      string
	SMTPPort      int
	EmailAccount  string
	EmailPassword string
}

var (
	mailerConfig  Config
	configInit    sync.Once
	isInitialized bool
)

func Initialize(cfg Config) error {
	var err error
	configInit.Do(func() {
		if cfg.EmailAccount == "" {
			err = fmt.Errorf("email account cannot be empty")
			return
		}
		if cfg.EmailPassword == "" {
			err = fmt.Errorf("email password cannot be empty")
			return
		}
		if cfg.SMTPHost == "" {
			err = fmt.Errorf("SMTP host cannot be empty")
			return
		}
		if cfg.SMTPPort == 0 {
			err = fmt.Errorf("SMTP port cannot be zero")
			return
		}

		mailerConfig = cfg
		isInitialized = true
		log.Println("Mailer initialized successfully")
	})
	return err
}

func HandleSendEmail(mailto string, subject string, bodyType string, body string) (string, error) {
	if !isInitialized {
		return "", fmt.Errorf("mailer not initialized. Call Initialize() first")
	}

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", mailerConfig.EmailAccount)
	mailer.SetHeader("To", mailto)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody(bodyType, body)

	dialer := gomail.NewDialer(
		mailerConfig.SMTPHost,
		mailerConfig.SMTPPort,
		mailerConfig.EmailAccount,
		mailerConfig.EmailPassword,
	)

	if err := dialer.DialAndSend(mailer); err != nil {
		log.Println("Error sending email:", err)
		return "", err
	}

	log.Println("Email sent successfully!")

	return "Email sent successfully!", nil
}

func SendEmailWithCC(mailto string, cc []string, subject string, bodyType string, body string) (string, error) {
	if !isInitialized {
		return "", fmt.Errorf("mailer not initialized. Call Initialize() first")
	}

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", mailerConfig.EmailAccount)
	mailer.SetHeader("To", mailto)
	if len(cc) > 0 {
		mailer.SetHeader("Cc", cc...)
	}
	mailer.SetHeader("Subject", subject)
	mailer.SetBody(bodyType, body)

	dialer := gomail.NewDialer(
		mailerConfig.SMTPHost,
		mailerConfig.SMTPPort,
		mailerConfig.EmailAccount,
		mailerConfig.EmailPassword,
	)

	if err := dialer.DialAndSend(mailer); err != nil {
		log.Println("Error sending email:", err)
		return "", err
	}

	log.Println("Email sent successfully!")

	return "Email sent successfully!", nil
}

func SendEmailWithAttachment(mailto string, subject string, bodyType string, body string, attachments []string) (string, error) {
	if !isInitialized {
		return "", fmt.Errorf("mailer not initialized. Call Initialize() first")
	}

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", mailerConfig.EmailAccount)
	mailer.SetHeader("To", mailto)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody(bodyType, body)

	for _, attachment := range attachments {
		mailer.Attach(attachment)
	}

	dialer := gomail.NewDialer(
		mailerConfig.SMTPHost,
		mailerConfig.SMTPPort,
		mailerConfig.EmailAccount,
		mailerConfig.EmailPassword,
	)

	if err := dialer.DialAndSend(mailer); err != nil {
		log.Println("Error sending email:", err)
		return "", err
	}

	log.Println("Email sent successfully!")

	return "Email sent successfully!", nil
}

// SendEmailWithMultipartFiles sends email with files from multipart form data
func SendEmailWithMultipartFiles(mailto string, subject string, bodyType string, body string, formFiles []*multipart.FileHeader) (string, error) {
	if !isInitialized {
		return "", fmt.Errorf("mailer not initialized. Call Initialize() first")
	}

	if len(formFiles) == 0 {
		return "", fmt.Errorf("no files provided")
	}

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", mailerConfig.EmailAccount)
	mailer.SetHeader("To", mailto)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody(bodyType, body)

	// Store temp file paths for cleanup
	tempFiles := []string{}

	// Attach each file from the multipart form
	for _, fileHeader := range formFiles {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error opening file %s: %v\n", fileHeader.Filename, err)
			cleanupTempFiles(tempFiles)
			return "", fmt.Errorf("failed to open file %s: %w", fileHeader.Filename, err)
		}

		// Create a temporary file to store the upload
		tmpFile, err := os.CreateTemp("", filepath.Base(fileHeader.Filename))
		if err != nil {
			log.Printf("Error creating temp file for %s: %v\n", fileHeader.Filename, err)
			file.Close()
			cleanupTempFiles(tempFiles)
			return "", fmt.Errorf("failed to create temp file: %w", err)
		}

		// Copy the uploaded file to the temporary file
		if _, err := io.Copy(tmpFile, file); err != nil {
			log.Printf("Error copying file %s: %v\n", fileHeader.Filename, err)
			file.Close()
			tmpFile.Close()
			cleanupTempFiles(tempFiles)
			return "", fmt.Errorf("failed to copy file %s: %w", fileHeader.Filename, err)
		}

		file.Close()
		tmpFile.Close()

		// Attach the temporary file
		mailer.Attach(tmpFile.Name())
		tempFiles = append(tempFiles, tmpFile.Name())
	}

	dialer := gomail.NewDialer(
		mailerConfig.SMTPHost,
		mailerConfig.SMTPPort,
		mailerConfig.EmailAccount,
		mailerConfig.EmailPassword,
	)

	if err := dialer.DialAndSend(mailer); err != nil {
		log.Println("Error sending email:", err)
		cleanupTempFiles(tempFiles)
		return "", err
	}

	// Clean up temporary files after sending
	cleanupTempFiles(tempFiles)

	log.Println("Email sent successfully with attachments!")

	return "Email sent successfully with attachments!", nil
}

// cleanupTempFiles removes temporary files
func cleanupTempFiles(files []string) {
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			log.Printf("Warning: failed to remove temp file %s: %v\n", f, err)
		}
	}
}
