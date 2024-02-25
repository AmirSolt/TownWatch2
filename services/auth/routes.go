package auth

import (
	"fmt"
	"net/http"
	"townwatch/base/basetemplates"

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
			eventId := sentry.CaptureException(fmt.Errorf("email not found on postFrom. ctx: %+v", ctx))
			err := fmt.Errorf("email missing from the form (%v)", eventId)
			basetemplates.Error(err).Render(ctx, ctx.Writer)
			return
		}
		err := auth.InitOTP(ctx, email)
		if err != nil {
			basetemplates.Error(err).Render(ctx, ctx.Writer)
			return
		}
		ctx.Redirect(http.StatusFound, "/join/verif")
	})

	auth.base.Engine.GET("/join/otp/:id", auth.RequireGuestMiddleware, func(ctx *gin.Context) {

		otpID := ctx.Param("id")
		errVOTP := auth.ValidateOTP(ctx, otpID)
		if errVOTP != nil {
			basetemplates.Error(errVOTP).Render(ctx, ctx.Writer)
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
			eventId := sentry.CaptureException(fmt.Errorf("email not found on postFrom. ctx: %+v", ctx))
			err := fmt.Errorf("email missing from the form (%v)", eventId)
			basetemplates.Error(err).Render(ctx, ctx.Writer)
			return
		}
		err := auth.DebugOTP(ctx, email)
		if err != nil {
			basetemplates.Error(err).Render(ctx, ctx.Writer)
			return
		}
		ctx.Redirect(http.StatusFound, "/")
	})

}
