package base

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
)

// Error tracking component

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
