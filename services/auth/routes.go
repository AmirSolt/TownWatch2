package auth

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

// routes:
// 1. signin
// 2. signout
// 3. signintest

func (auth *Auth) registerAuthRoutes() {
	auth.authRoutes()

	if !auth.base.IS_PROD {
		auth.authTestRoutes()
	}
}

func (auth *Auth) authRoutes() {

	auth.base.Engine.POST("/join/signin", auth.RequireGuestMiddleware, func(ctx *gin.Context) {

		email := ctx.PostForm("email")
		if email == "" {
			errDev := fmt.Errorf("email not found on postFrom. ctx: %+v", ctx)
			eventId := sentry.CaptureException(errDev)
			errUser := fmt.Errorf("email missing from the form (%s)", *eventId)
			ctx.String(http.StatusBadRequest, errUser.Error())
			return
		}
		errComm := auth.InitOTP(ctx, email)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}
		ctx.Redirect(http.StatusFound, "/join/verif")
	})

	auth.base.Engine.GET("/join/otp/:id", auth.RequireGuestMiddleware, func(ctx *gin.Context) {

		otpID := ctx.Param("id")
		errComm := auth.ValidateOTP(ctx, otpID)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}
		ctx.Redirect(http.StatusFound, "/")
	})

	auth.base.Engine.GET("/join/signout", auth.RequireUserMiddleware, func(ctx *gin.Context) {
		auth.Signout(ctx)

		ctx.Redirect(http.StatusFound, "/")
	})
}

func (auth *Auth) authTestRoutes() {

	auth.base.Engine.POST("/join/signin/debug", auth.RequireGuestMiddleware, func(ctx *gin.Context) {

		email := ctx.PostForm("email")
		if email == "" {
			errDev := fmt.Errorf("email not found on postFrom. ctx: %+v", ctx)
			eventId := sentry.CaptureException(errDev)
			errUser := fmt.Errorf("email missing from the form (%s)", *eventId)
			ctx.String(http.StatusBadRequest, errUser.Error())
			return
		}
		errComm := auth.DebugOTP(ctx, email)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		ctx.Redirect(http.StatusFound, "/")
	})

}
