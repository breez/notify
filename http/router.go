package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/breez/notify/channel"
	"github.com/breez/notify/config"
	"github.com/breez/notify/notify"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/martian/v3/log"
)

type MobilePushWebHookQuery struct {
	Platform string  `form:"platform" binding:"required,oneof=ios android"`
	Token    string  `form:"token" binding:"required"`
	AppData  *string `form:"app_data"`
}

type NotificationConvertible interface {
	RequiresCallback() bool
	ToNotification(query *MobilePushWebHookQuery) *notify.Notification
}

type LnurlPayInfoPayload struct {
	Template string `json:"template" binding:"required,eq=lnurlpay_info"`
	Data     struct {
		CallbackURL string `json:"callback_url" binding:"required"`
		ReplyURL    string `json:"reply_url" binding:"required"`
	} `json:"data"`
}

func (p *LnurlPayInfoPayload) RequiresCallback() bool {
	return false
}

func (p *LnurlPayInfoPayload) ToNotification(query *MobilePushWebHookQuery) *notify.Notification {
	return &notify.Notification{
		Template:         p.Template,
		DisplayMessage:   "Receiving payment",
		Type:             query.Platform,
		TargetIdentifier: query.Token,
		AppData:          query.AppData,
		Data: map[string]interface{}{
			"callback_url": p.Data.CallbackURL,
			"reply_url":    p.Data.ReplyURL,
		},
	}
}

type LnurlPayInvoicePayload struct {
	Template string `json:"template" binding:"required,eq=lnurlpay_invoice"`
	Data     struct {
		Amount    uint64  `json:"amount" binding:"required,min=1"`
		Comment   *string `json:"comment"`
		ReplyURL  string  `json:"reply_url" binding:"required"`
		VerifyURL *string `json:"verify_url"`
	} `json:"data"`
}

func (p *LnurlPayInvoicePayload) RequiresCallback() bool {
	return false
}

func (p *LnurlPayInvoicePayload) ToNotification(query *MobilePushWebHookQuery) *notify.Notification {
	notification := notify.Notification{
		Template:         p.Template,
		DisplayMessage:   "Invoice requested",
		Type:             query.Platform,
		TargetIdentifier: query.Token,
		AppData:          query.AppData,
		Data: map[string]interface{}{
			"amount":    p.Data.Amount,
			"reply_url": p.Data.ReplyURL,
		},
	}
	if p.Data.Comment != nil {
		notification.Data["comment"] = p.Data.Comment
	}
	if p.Data.VerifyURL != nil {
		notification.Data["verify_url"] = p.Data.VerifyURL
	}

	return &notification
}

type LnurlPayVerifyPayload struct {
	Template string `json:"template" binding:"required,eq=lnurlpay_verify"`
	Data     struct {
		PaymentHash string `json:"payment_hash" binding:"required"`
		ReplyURL    string `json:"reply_url" binding:"required"`
	} `json:"data"`
}

func (p *LnurlPayVerifyPayload) RequiresCallback() bool {
	return false
}

func (p *LnurlPayVerifyPayload) ToNotification(query *MobilePushWebHookQuery) *notify.Notification {
	return &notify.Notification{
		Template:         p.Template,
		DisplayMessage:   "Verify payment",
		Type:             query.Platform,
		TargetIdentifier: query.Token,
		AppData:          query.AppData,
		Data: map[string]interface{}{
			"payment_hash": p.Data.PaymentHash,
			"reply_url":    p.Data.ReplyURL,
		},
	}
}

type PaymentReceivedPayload struct {
	Template string `json:"template" binding:"required,eq=payment_received"`
	Data     struct {
		PaymentHash string `json:"payment_hash" binding:"required"`
	} `json:"data"`
}

func (p *PaymentReceivedPayload) RequiresCallback() bool {
	return false
}

func (p *PaymentReceivedPayload) ToNotification(query *MobilePushWebHookQuery) *notify.Notification {
	return &notify.Notification{
		Template:         p.Template,
		DisplayMessage:   "Incoming payment",
		Type:             query.Platform,
		TargetIdentifier: query.Token,
		AppData:          query.AppData,
		Data:             map[string]interface{}{"payment_hash": p.Data.PaymentHash},
	}
}

type TxConfirmedPayload struct {
	Template string `json:"template" binding:"required,eq=tx_confirmed"`
	Data     struct {
		TxID string `json:"tx_id" binding:"required"`
	} `json:"data"`
}

func (p *TxConfirmedPayload) RequiresCallback() bool {
	return false
}

func (p *TxConfirmedPayload) ToNotification(query *MobilePushWebHookQuery) *notify.Notification {
	return &notify.Notification{
		Template:         p.Template,
		DisplayMessage:   "Transaction confirmed",
		Type:             query.Platform,
		TargetIdentifier: query.Token,
		AppData:          query.AppData,
		Data:             map[string]interface{}{"tx_id": p.Data.TxID},
	}
}

type AddressTxsConfirmedPayload struct {
	Template string `json:"template" binding:"required,eq=address_txs_confirmed"`
	Data     struct {
		Address string `json:"address" binding:"required"`
	} `json:"data"`
}

func (p *AddressTxsConfirmedPayload) RequiresCallback() bool {
	return false
}

func (p *AddressTxsConfirmedPayload) ToNotification(query *MobilePushWebHookQuery) *notify.Notification {
	return &notify.Notification{
		Template:         p.Template,
		DisplayMessage:   "Address transactions confirmed",
		Type:             query.Platform,
		TargetIdentifier: query.Token,
		AppData:          query.AppData,
		Data:             map[string]interface{}{"address": p.Data.Address},
	}
}

type SwapUpdatedPayload struct {
	Event string `json:"event" binding:"required,eq=swap.update"`
	Data  struct {
		Id     string `json:"id" binding:"required"`
		Status string `json:"status" binding:"required"`
	} `json:"data"`
}

func (p *SwapUpdatedPayload) RequiresCallback() bool {
	return false
}

func (p *SwapUpdatedPayload) ToNotification(query *MobilePushWebHookQuery) *notify.Notification {
	return &notify.Notification{
		Template:         notify.NOTIFICATION_SWAP_UPDATED,
		DisplayMessage:   "Swap updated",
		Type:             query.Platform,
		TargetIdentifier: query.Token,
		AppData:          query.AppData,
		Data:             map[string]interface{}{"id": p.Data.Id, "status": p.Data.Status},
	}
}

type InvoiceRequestPayload struct {
	Event string `json:"event" binding:"required,eq=invoice.request"`
	Data  struct {
		Offer          string `json:"offer" binding:"required"`
		InvoiceRequest string `json:"invoiceRequest" binding:"required"`
	} `json:"data"`
}

func (p *InvoiceRequestPayload) RequiresCallback() bool {
	return true
}

func (p *InvoiceRequestPayload) ToNotification(query *MobilePushWebHookQuery) *notify.Notification {
	return &notify.Notification{
		Template:         notify.NOTIFICATION_INVOICE_REQUEST,
		DisplayMessage:   "Invoice request",
		Type:             query.Platform,
		TargetIdentifier: query.Token,
		AppData:          query.AppData,
		Data:             map[string]interface{}{"offer": p.Data.Offer, "invoice_request": p.Data.InvoiceRequest},
	}
}

type NwcEventPayload struct {
	Template  string `json:"template" binding:"required,eq=nwc_event"`
	Data      struct {
		EventID string `json:"event_id" binding:"required"`
	} `json:"data"`
}

func (p *NwcEventPayload) RequiresCallback() bool {
	return false
}

func (p *NwcEventPayload) ToNotification(query *MobilePushWebHookQuery) *notify.Notification {
	return &notify.Notification{
		Template:         notify.NOTIFICATION_NWC_EVENT,
		DisplayMessage:   "NWC Event Received",
		Type:             query.Platform,
		TargetIdentifier: query.Token,
		AppData:          query.AppData,
		Data:             map[string]interface{}{"event_id": p.Data.EventID},
	}
}

func Run(notifier *notify.Notifier, channel *channel.HttpCallbackChannel, config *config.HTTPConfig) error {
	r := setupRouter(notifier, channel)
	r.SetTrustedProxies(nil)
	return r.Run(config.Address)
}

func setupRouter(notifier *notify.Notifier, channel *channel.HttpCallbackChannel) *gin.Engine {
	r := gin.Default()
	router := r.Group("api/v1")
	addRouter(router, notifier, channel)
	return r
}

func addRouter(r *gin.RouterGroup, notifier *notify.Notifier, channel *channel.HttpCallbackChannel) {
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
			&LnurlPayVerifyPayload{},
			&SwapUpdatedPayload{},
			&InvoiceRequestPayload{},
			&NwcEventPayload{},
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
			response, err := channel.Notify(c, notifier, r.BasePath(), validPayload.ToNotification(&query))
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
			if err := notifier.Notify(c, validPayload.ToNotification(&query)); err != nil {
				log.Debugf("failed to notify, query: %v, error: %v", query, err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.Status(http.StatusOK)
	})

	r.POST("/response/:responseId", func(c *gin.Context) {
		responseId := c.Param("responseId")

		reqId, err := strconv.ParseUint(responseId, 10, 64)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.New("invalid response"))
			return
		}

		all, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New("internal error"))
			return
		}

		if err := channel.OnResponse(reqId, string(all)); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	})
}
