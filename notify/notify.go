package notify

import (
	"context"
	"errors"

	"github.com/breez/notify/config"
	"github.com/golang-queue/queue"
	"github.com/google/martian/v3/log"
)

const (
	NOTIFICATION_PAYMENT_RECEIVED    = "payment_received"
	NOTIFICATION_TX_CONFIRMED        = "tx_confirmed"
	NOTIFICATION_ADDRESS_TXS_CHANGED = "address_txs_changed"
	NOTIFICATION_WEBHOOK_CALLBACK    = "webhook_callback_message"
)

var (
	ErrServiceNotFound = errors.New("Service not found")
)

type Notification struct {
	Template         string
	DisplayMessage   string
	Type             string
	TargetIdentifier string
	Data             map[string]string
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
			log.Errorf("could not find service %+v %v", request.Type)
			return ErrServiceNotFound
		}
		if err := service.Send(c, request); err != nil {
			log.Errorf("failed to send notification %+v %v", request, err)
			return err
		}
		log.Infof("succeed to send notification %+v", request)
		return nil
	})
}
