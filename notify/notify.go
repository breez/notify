package notify

import (
	"context"
	"errors"

	"github.com/breez/notify/config"
	"github.com/golang-queue/queue"
)

const (
	NOTIFICATION_PAYMENT_RECEIVED   = "payment_received"
	NOTIFICATION_LSP_CLOSES_CHANNEL = "lsp_closes_channel"
)

var (
	ErrServiceNotFound = errors.New("Service not found")
)

type Notification struct {
	Template string
	Type     string
	Token    string
}

type Service interface {
	Send(context context.Context, req *Notification) error
}

type Notifier struct {
	queue         *queue.Queue
	serviceByType map[string]Service
}

func NewNotifier(
	config *config.Config,
	services map[string]Service) *Notifier {
	q := queue.NewPool(config.WorkersNum)
	return &Notifier{
		queue:         q,
		serviceByType: services,
	}
}

func (n *Notifier) Notify(c context.Context, request *Notification) error {
	return n.queue.QueueTask(func(ctx context.Context) error {
		service, ok := n.serviceByType[request.Type]
		if !ok {
			return ErrServiceNotFound
		}
		return service.Send(c, request)
	})
}
