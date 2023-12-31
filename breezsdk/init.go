package breezsdk

import (
	"firebase.google.com/go/messaging"
	"github.com/breez/notify/config"
	"github.com/breez/notify/notify"
	"github.com/breez/notify/notify/services"
)

func NewNotifier(c *config.Config, fcmClient *messaging.Client) (*notify.Notifier, error) {
	fcm := services.NewFCM(createMessageFactory(), fcmClient)
	return notify.NewNotifier(c, map[string]notify.Service{
		"ios":     fcm,
		"android": fcm,
	}), nil
}

func createMessageFactory() services.FCMMessageBuilder {
	return func(notification *notify.Notification) (*messaging.Message, error) {

		switch notification.Template {
		case notify.NOTIFICATION_PAYMENT_RECEIVED,
			notify.NOTIFICATION_TX_CONFIRMED,
			notify.NOTIFICATION_ADDRESS_TXS_CHANGED,
			notify.NOTIFICATION_WEBHOOK_CALLBACK:

			return createSilentPush(notification)
		}

		return nil, nil
	}
}

func createSilentPush(notification *notify.Notification) (*messaging.Message, error) {
	data := notification.Data
	data["notification_type"] = notification.Template

	return &messaging.Message{
		Token: notification.TargetIdentifier,
		Data:  data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					ContentAvailable: false,
					MutableContent:   true,
				},
			},
		},
	}, nil
}
