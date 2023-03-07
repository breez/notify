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
		"fcm": fcm,
	}), nil
}

func createMessageFactory() services.FCMMessageFactory {
	return func(notification *notify.Notification) *messaging.Message {
		switch notification.Template {
		case notify.NOTIFICATION_PAYMENT_RECEIVED:
		case notify.NOTIFICATION_LSP_CLOSES_CHANNEL:
			return createSilentPush(notification)
		}

		return nil
	}
}

func createSilentPush(notification *notify.Notification) *messaging.Message {
	return &messaging.Message{
		Token: notification.Token,
		Data: map[string]string{
			"notification_type": notification.Template,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					ContentAvailable: true,
					MutableContent:   true,
				},
			},
		},
	}
}
