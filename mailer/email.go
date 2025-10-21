package mailer

import (
	"fmt"
	"log"
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
