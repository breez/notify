package breezsdk

import (
	"encoding/json"
	"fmt"

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
			notify.NOTIFICATION_ADDRESS_TXS_CONFIRMED,
			notify.NOTIFICATION_LNURLPAY_INFO,
			notify.NOTIFICATION_LNURLPAY_INVOICE,
			notify.NOTIFICATION_SWAP_UPDATED:

			return createPush(notification)
		}

		return nil, nil
	}
}

func createPush(notification *notify.Notification) (*messaging.Message, error) {
	data := make(map[string]string)

	data["notification_type"] = notification.Template
	if notification.AppData != nil {
		data["app_data"] = *notification.AppData
	}
	payload, err := json.Marshal(notification.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal notification data %v", err)
	}
	data["notification_payload"] = string(payload)

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
					Alert: &messaging.ApsAlert{
						Title: notification.DisplayMessage,
					},
					ContentAvailable: false,
					MutableContent:   true,
				},
			},
		},
	}, nil
}
