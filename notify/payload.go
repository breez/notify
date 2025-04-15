package notify

import "github.com/breez/notify/notification"

type MobilePushWebHookQuery struct {
	Platform string  `form:"platform" binding:"required,oneof=ios android"`
	Token    string  `form:"token" binding:"required"`
	AppData  *string `form:"app_data"`
}

type NotificationConvertible interface {
	RequiresCallback() bool
	ToNotification(query *MobilePushWebHookQuery) *notification.Notification
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

func (p *LnurlPayInfoPayload) ToNotification(query *MobilePushWebHookQuery) *notification.Notification {
	return &notification.Notification{
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
		Amount   uint64 `json:"amount" binding:"required,min=1"`
		ReplyURL string `json:"reply_url" binding:"required"`
	} `json:"data"`
}

func (p *LnurlPayInvoicePayload) RequiresCallback() bool {
	return false
}

func (p *LnurlPayInvoicePayload) ToNotification(query *MobilePushWebHookQuery) *notification.Notification {
	return &notification.Notification{
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

func (p *PaymentReceivedPayload) ToNotification(query *MobilePushWebHookQuery) *notification.Notification {
	return &notification.Notification{
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

func (p *TxConfirmedPayload) ToNotification(query *MobilePushWebHookQuery) *notification.Notification {
	return &notification.Notification{
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

func (p *AddressTxsConfirmedPayload) ToNotification(query *MobilePushWebHookQuery) *notification.Notification {
	return &notification.Notification{
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

func (p *SwapUpdatedPayload) ToNotification(query *MobilePushWebHookQuery) *notification.Notification {
	return &notification.Notification{
		Template:         NOTIFICATION_SWAP_UPDATED,
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
		Offer     string `json:"offer" binding:"required"`
		InvoiceRequest string `json:"invoiceRequest" binding:"required"`
	} `json:"data"`
}

func (p *InvoiceRequestPayload) RequiresCallback() bool {
	return true
}

func (p *InvoiceRequestPayload) ToNotification(query *MobilePushWebHookQuery) *notification.Notification {
	return &notification.Notification{
		Template:         NOTIFICATION_INVOICE_REQUEST,
		DisplayMessage:   "Invoice request",
		Type:             query.Platform,
		TargetIdentifier: query.Token,
		AppData:          query.AppData,
		Data:             map[string]interface{}{"offer": p.Data.Offer, "invoice_request": p.Data.InvoiceRequest},
	}
}
