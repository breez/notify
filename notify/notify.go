package notify

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/breez/notify/channel"
	"github.com/breez/notify/notification"
	"github.com/breez/notify/config"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-queue/queue"
	"github.com/google/martian/v3/log"
)

const (
	NOTIFICATION_PAYMENT_RECEIVED      = "payment_received"
	NOTIFICATION_TX_CONFIRMED          = "tx_confirmed"
	NOTIFICATION_ADDRESS_TXS_CONFIRMED = "address_txs_confirmed"
	NOTIFICATION_LNURLPAY_INFO         = "lnurlpay_info"
	NOTIFICATION_LNURLPAY_INVOICE      = "lnurlpay_invoice"
	NOTIFICATION_SWAP_UPDATED          = "swap_updated"
	NOTIFICATION_INVOICE_REQUEST       = "invoice_request"
)

var (
	ErrServiceNotFound = errors.New("Service not found")
)

type Service interface {
	Send(context context.Context, req *notification.Notification) error
}

type Notifier struct {
	channel       channel.WebhookChannel
	queue         *queue.Queue
	serviceByType map[string]Service
}

func NewNotifier( config *config.Config, services map[string]Service) *Notifier {
	q := queue.NewPool(config.WorkersNum)
	channel := channel.NewHttpCallbackChannel(config.ExternalURL)
	return &Notifier{
		channel:       channel,
		queue:         q,
		serviceByType: services,
	}
}

func (n *Notifier) AddRouter(r *gin.RouterGroup) {
	n.addRouter(r)
	n.channel.AddRouter(r)
}

func (n *Notifier) Notify(c context.Context, request *notification.Notification) error {
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

func (n *Notifier) addRouter(r *gin.RouterGroup) {
	r.POST("/notify", func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Make sure the query string fits the mobile push structure
		var query MobilePushWebHookQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		// Find a matching notification payload
		payloads := []NotificationConvertible{
			&PaymentReceivedPayload{},
			&TxConfirmedPayload{},
			&AddressTxsConfirmedPayload{},
			&LnurlPayInfoPayload{},
			&LnurlPayInvoicePayload{},
			&SwapUpdatedPayload{},
			&InvoiceRequestPayload{},
		}
		var validPayload NotificationConvertible
		for _, p := range payloads {
			if err := c.ShouldBindBodyWith(p, binding.JSON); err != nil {
				continue
			}
			validPayload = p
			break
		}

		if validPayload == nil {
			log.Debugf("invalid payload, body: %s", body)
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("unsupported payload, body: %s", body))
			return
		}

		if validPayload.RequiresCallback() {
			response, err := n.channel.Notify(c, n, validPayload.ToNotification(&query))
			if c.IsAborted() {
				return
			}
			if err != nil {
				log.Debugf("failed to notify with channel, query: %v, error: %v", query, err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			c.Header("Content-Type", "application/json")
			c.Writer.Write([]byte(response))
			return
		} else {
			if err := n.Notify(c, validPayload.ToNotification(&query)); err != nil {
				log.Debugf("failed to notify, query: %v, error: %v", query, err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusOK)
	})
}
