package channel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/breez/notify/notification"
	"github.com/gin-gonic/gin"
	"github.com/google/martian/v3/log"
)

const (
	callbackTimeout = 60 * time.Second
)

type ChannelNotifier interface {
	Notify(c context.Context, request *notification.Notification) error
}

type WebhookChannel interface {
	AddRouter(r *gin.RouterGroup)
	Notify(c context.Context, n ChannelNotifier, request *notification.Notification) (string, error)
}

type PendingRequest struct {
	id     uint64
	result chan string
}

type HttpCallbackChannel struct {
	sync.Mutex
	httpClient      *http.Client
	callbackBaseURL string
	basePath        string
	random          *rand.Rand
	pendingRequests map[uint64]*PendingRequest
}

func NewHttpCallbackChannel(callbackBaseURL string) *HttpCallbackChannel {
	channel := &HttpCallbackChannel{
		httpClient:      http.DefaultClient,
		callbackBaseURL: callbackBaseURL,
		basePath:        "",
		random:          rand.New(rand.NewSource(time.Now().UnixNano())),
		pendingRequests: make(map[uint64]*PendingRequest),
	}

	return channel
}

func (p *HttpCallbackChannel) AddRouter(r *gin.RouterGroup) {
	p.basePath = r.BasePath()
	p.addRouter(r)
}

func (p *HttpCallbackChannel) Notify(c context.Context, n ChannelNotifier, request *notification.Notification) (string, error) {
	reqID := p.random.Uint64()
	callbackURL := fmt.Sprintf("%s/%s/response/%d", p.callbackBaseURL, p.basePath, reqID)
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

	if err := n.Notify(c, request); err != nil {
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

func (p *HttpCallbackChannel) addRouter(r *gin.RouterGroup) {
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

		if err := p.onResponse(reqId, string(all)); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	})
}

func (p *HttpCallbackChannel) onResponse(reqID uint64, payload string) error {
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
