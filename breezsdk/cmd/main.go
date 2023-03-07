package main

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"github.com/Netflix/go-env"

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
	firebaseApp, err := firebase.NewApp(context.Background(), nil)
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
