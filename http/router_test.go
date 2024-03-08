package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/breez/notify/config"
	"github.com/breez/notify/notify"
	"gotest.tools/assert"
)

func TestPaymentReceivedHook(t *testing.T) {
	testAppData := "testdata"
	query := MobilePushWebHookQuery{
		Platform: "android",
		Token:    "1234",
		AppData:  &testAppData,
	}

	paymentReceivedPayload := PaymentReceivedPayload{
		Template: notify.NOTIFICATION_PAYMENT_RECEIVED,
		Data: struct {
			PaymentHash string "json:\"payment_hash\" binding:\"required\""
		}{
			PaymentHash: "1234",
		},
	}

	body, err := json.Marshal(paymentReceivedPayload)
	if err != nil {
		t.Fatalf("failed to marshal notification %v", err)
	}
	expected := paymentReceivedPayload.ToNotification(&query)
	testValidNotification(t, "/api/v1/notify?platform=android&token=1234&app_data=testdata", body, expected)
}

func TestTxConfirmedHook(t *testing.T) {
	query := MobilePushWebHookQuery{
		Platform: "android",
		Token:    "1234",
	}
	txConfirmedPayload := TxConfirmedPayload{
		Template: notify.NOTIFICATION_TX_CONFIRMED,
		Data: struct {
			TxID string "json:\"tx_id\" binding:\"required\""
		}{
			TxID: "1234",
		},
	}
	body, err := json.Marshal(txConfirmedPayload)
	if err != nil {
		t.Fatalf("failed to marshal notification %v", err)
	}
	expected := txConfirmedPayload.ToNotification(&query)
	testValidNotification(t, "/api/v1/notify?platform=android&token=1234", body, expected)
}

func TestAddressTXsChangedHook(t *testing.T) {
	query := MobilePushWebHookQuery{
		Platform: "android",
		Token:    "1234",
	}
	txAddressTxsConfirmedPayload := AddressTxsConfirmedPayload{
		Template: notify.NOTIFICATION_ADDRESS_TXS_CONFIRMED,
		Data: struct {
			Address string "json:\"address\" binding:\"required\""
		}{
			Address: "1234",
		},
	}
	body, err := json.Marshal(txAddressTxsConfirmedPayload)
	if err != nil {
		t.Fatalf("failed to marshal notification %v", err)
	}
	expected := txAddressTxsConfirmedPayload.ToNotification(&query)
	testValidNotification(t, "/api/v1/notify?platform=android&token=1234", body, expected)
}

func testValidNotification(t *testing.T, url string, body []byte, expected *notify.Notification) {
	service := newTestService()
	config := &config.Config{WorkersNum: 2}
	notifier := notify.NewNotifier(config, map[string]notify.Service{"android": service})

	router := setupRouter(notifier)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.DeepEqual(t, *expected, *<-service.sentQueue)
}

type TestService struct {
	sentQueue chan *notify.Notification
}

func newTestService() *TestService {
	queue := make(chan *notify.Notification, 10)
	return &TestService{sentQueue: queue}
}

func (t *TestService) Send(c context.Context, notification *notify.Notification) error {
	t.sentQueue <- notification
	return nil
}
