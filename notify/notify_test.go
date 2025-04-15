package notify

import (
	"context"
	"testing"

	"github.com/breez/notify/config"
	"github.com/breez/notify/notification"
	"gotest.tools/v3/assert"
)

type TestService struct {
	sentQueue chan *notification.Notification
}

func newTestService() *TestService {
	queue := make(chan *notification.Notification, 10)
	return &TestService{sentQueue: queue}
}

func (t *TestService) Send(c context.Context, notification *notification.Notification) error {
	t.sentQueue <- notification
	return nil
}

func TestNotify(t *testing.T) {
	service := newTestService()
	config := &config.Config{WorkersNum: 2}
	notifier := NewNotifier(config, map[string]Service{"test": service})
	n := notification.Notification{
		Template:         "t1",
		Type:             "test",
		TargetIdentifier: "token1",
	}
	notifier.Notify(context.Background(), &n)

	var notifications []notification.Notification
	res := <-service.sentQueue
	notifications = append(notifications, *res)
	assert.Assert(t, len(notifications) == 1)
	assert.DeepEqual(t, notifications[0], n)
}
