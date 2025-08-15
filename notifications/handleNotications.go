package notifications

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

func initializeFirebaseApp() (*messaging.Client, error) {
	opt := option.WithCredentialsFile("adminsdk.json")
	config := &firebase.Config{ProjectID: "test-dashboard-65d9c"}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Println("error initializing firebase app: ", err)
		return nil, err
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Println("error initializing firebase ##  messaging client: ", err)
		return nil, err
	}

	return client, nil
}

func SendNotification(deviceToken, title, body string) error {
	client, err := initializeFirebaseApp()
	if err != nil {
		return err
	}

	message := &messaging.Message{
		Token: deviceToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	_, err = client.Send(context.Background(), message)
	if err != nil {
		log.Printf("Error sending notification: %v %v", err, deviceToken)
		return err
	}

	return nil
}
