package services

import (
	"context"
	"errors"
	"fmt"

	"firebase.google.com/go/messaging"
	"github.com/breez/notify/notification"
)

var (
	ErrUnrecognizedTemplate = errors.New("unrecognized template")
)

type FCMMessageBuilder func(req *notification.Notification) (*messaging.Message, error)
type FCM struct {
	messageBuilder FCMMessageBuilder
	client         *messaging.Client
}

func NewFCM(messageBuilder FCMMessageBuilder, client *messaging.Client) *FCM {
	return &FCM{messageBuilder: messageBuilder, client: client}
}

func (f *FCM) Send(context context.Context, req *notification.Notification) error {
	pushNotification, err := f.messageBuilder(req)
	if err != nil {
		return fmt.Errorf("failed to create message %v", err)
	}
	if pushNotification == nil {
		return ErrUnrecognizedTemplate
	}
	_, err = f.client.Send(context, pushNotification)
	if err != nil {
		return fmt.Errorf("failed to send fcm message %v", err)
	}

	return nil
}
