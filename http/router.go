package http

import (
	"log"
	"net/http"

	"github.com/breez/notify/notify"
	"github.com/gin-gonic/gin"
)

type WebHookQuery struct {
	Template string
	Type     string
	Token    string
}

func Run(notifier *notify.Notifier) error {
	r := gin.Default()
	router := r.Group("api/v1")

	addWebHookRouter(router, notifier)
	return r.Run()
}

func addWebHookRouter(r *gin.RouterGroup, notifier *notify.Notifier) {
	r.POST("/notify", func(c *gin.Context) {
		var query WebHookQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if err := notifier.Notify(c, &notify.Notification{
			Template: query.Template,
			Type:     query.Type,
			Token:    query.Token,
		}); err != nil {
			log.Printf("failed to notify, query: %v, error: %v", query, err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)
	})
}
