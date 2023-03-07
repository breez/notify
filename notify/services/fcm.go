package services

import (
	"context"
	"errors"
	"fmt"

	"firebase.google.com/go/messaging"
	"github.com/breez/notify/notify"
)

var (
	ErrUnrecognizedTemplate = errors.New("unrecognized template")
)

type FCMMessageFactory func(req *notify.Notification) *messaging.Message
type FCM struct {
	messageFactory FCMMessageFactory
	client         *messaging.Client
}

func NewFCM(messageFactory FCMMessageFactory, client *messaging.Client) *FCM {
	return &FCM{messageFactory: messageFactory, client: client}
}

func (f *FCM) Send(context context.Context, req *notify.Notification) error {
	pushNotification := f.messageFactory(req)
	if pushNotification == nil {
		return ErrUnrecognizedTemplate
	}
	_, err := f.client.Send(context, pushNotification)
	if err != nil {
		return fmt.Errorf("failed to send fcm message %v", err)
	}

	return nil
}
