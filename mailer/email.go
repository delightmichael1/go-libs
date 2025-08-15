package mailer

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

func init() {
	if os.Getenv("ENVIRONMENT") == "" || os.Getenv("ENVIRONMENT") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}
}

func HandleSendEmail(mailto string, subject string, bodyType string, body string) (string, error) {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", os.Getenv("GMAIL_ACCOUNT"))
	mailer.SetHeader("To", mailto)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody(bodyType, body)

	dialer := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("GMAIL_ACCOUNT"), os.Getenv("GMAIL_APP_PASSWORD"))

	if err := dialer.DialAndSend(mailer); err != nil {
		log.Println("Error sending email:", err)
		return "", err
	}

	log.Println("Email sent successfully!")

	return "Email sent successfully!", nil
}
