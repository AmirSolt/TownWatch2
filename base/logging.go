package base

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

type ErrorComm struct {
	*sentry.EventID
	UserMsg error
	DevMsg  error
}

func (base *Base) loadLogging() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              base.GLITCHTIP_DSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Debug:            !base.IS_PROD,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	base.Engine.Use(sentrygin.New(sentrygin.Options{}))

}

func (base *Base) killLogging() {
	sentry.Flush(time.Second * 5)
}

func (base *Base) HandleRouteRenderError(ctx *gin.Context, errRender error) {
	if errRender != nil {
		eventId := sentry.CaptureException(errRender)
		errComm := &ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("there was an error rendering this page. You can use this Event ID to report the problem (%s)", *eventId),
			DevMsg:  errRender,
		}
		ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
	}
}
