package channel

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/breez/notify/notify"
	"github.com/google/martian/v3/log"
)

const (
	callbackTimeout = 60 * time.Second
)

type PendingRequest struct {
	id     uint64
	result chan string
}

type HttpCallbackChannel struct {
	sync.Mutex
	httpClient      *http.Client
	callbackBaseURL string
	random          *rand.Rand
	pendingRequests map[uint64]*PendingRequest
}

func NewHttpCallbackChannel(callbackBaseURL string) *HttpCallbackChannel {
	channel := &HttpCallbackChannel{
		httpClient:      http.DefaultClient,
		callbackBaseURL: strings.TrimRight(callbackBaseURL, "/"),
		random:          rand.New(rand.NewSource(time.Now().UnixNano())),
		pendingRequests: make(map[uint64]*PendingRequest),
	}

	return channel
}

func (p *HttpCallbackChannel) Notify(c context.Context, notifier *notify.Notifier, basePath string, request *notify.Notification) (string, error) {
	reqID := p.random.Uint64()
	trimmedBasePath := strings.Trim(basePath, "/")
	callbackURL := fmt.Sprintf("%s/%s/response/%d", p.callbackBaseURL, trimmedBasePath, reqID)
	request.Data["reply_url"] = callbackURL

	pendingRequest := &PendingRequest{
		id:     reqID,
		result: make(chan string, 1),
	}
	p.Lock()
	p.pendingRequests[reqID] = pendingRequest
	p.Unlock()

	// We only delete the request from the map and close the channel only if it was not deleted before.
	defer func() {
		p.Lock()
		req, ok := p.pendingRequests[reqID]
		if ok {
			p.deleteRequestAndClose(req)
		}
		p.Unlock()
	}()

	log.Debugf("waiting for response: %v", callbackURL)

	if err := notifier.Notify(c, request); err != nil {
		log.Debugf("failed to notify, request: %v, error: %v", request, err)
		return "", err
	}

	select {
	case result := <-pendingRequest.result:
		return result, nil
	case <-c.Done():
		return "", errors.New("canceled")
	case <-time.After(callbackTimeout):
		return "", errors.New("timeout")
	}
}

func (p *HttpCallbackChannel) OnResponse(reqID uint64, payload string) error {
	p.Lock()
	defer p.Unlock()
	pendingRequest, ok := p.pendingRequests[reqID]
	if !ok {
		return errors.New("unknown request id")
	}
	pendingRequest.result <- payload
	// We only delete the request from the map and close the channel.
	p.deleteRequestAndClose(pendingRequest)
	return nil
}

func (p *HttpCallbackChannel) deleteRequestAndClose(req *PendingRequest) {
	delete(p.pendingRequests, req.id)
	close(req.result)
}
