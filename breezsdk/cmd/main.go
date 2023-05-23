package main

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"github.com/Netflix/go-env"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	//"github.com/breez/notify/breezsdk"
	"github.com/breez/notify/breezsdk"
	"github.com/breez/notify/config"
	"github.com/breez/notify/http"
)

func main() {
	var config config.Config
	if _, err := env.UnmarshalFromEnviron(&config); err != nil {
		log.Fatalf("failed to load config %v", err)
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("failed to validate config %v", err)
	}

	creds, err := google.CredentialsFromJSON(context.Background(), []byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")), "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		log.Fatalf("failed to get google credentials %v", err)
	}
	firebaseApp, err := firebase.NewApp(context.Background(), nil, option.WithCredentials(creds))
	if err != nil {
		log.Fatalf("failed to create firebase application %v", err)
	}
	fcmMessaging, err := firebaseApp.Messaging(context.Background())
	if err != nil {
		log.Fatalf("failed to create firebase messaging %v", err)
	}
	notifier, err := breezsdk.NewNotifier(&config, fcmMessaging)
	if err != nil {
		log.Fatalf("failed to create breezsdk notifier %v", err)
	}
	if err := http.Run(notifier, &config.HTTPConfig); err != nil {
		log.Printf("web server has exited with error")
	}
}
