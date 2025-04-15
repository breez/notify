package http

import (
	"github.com/breez/notify/config"
	"github.com/breez/notify/notify"
	"github.com/gin-gonic/gin"
)

func Run(notifier *notify.Notifier, config *config.HTTPConfig) error {
	r := setupRouter(notifier)
	r.SetTrustedProxies(nil)
	return r.Run(config.Address)
}

func setupRouter(notifier *notify.Notifier) *gin.Engine {
	r := gin.Default()
	router := r.Group("api/v1")
	notifier.AddRouter(router)
	return r
}
