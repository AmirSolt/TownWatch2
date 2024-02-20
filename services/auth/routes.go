package auth

import (
	"errors"
	"net/http"
	authtemplates "townwatch/base/basetemplates"

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
			err := errors.New("error: No email found")
			authtemplates.Error(err).Render(ctx, ctx.Writer)
			return
		}
		err := auth.InitOTP(ctx, email)
		if err != nil {
			authtemplates.Error(err).Render(ctx, ctx.Writer)
			return
		}
		ctx.Redirect(http.StatusFound, "/join/verif")
	})

	auth.base.Engine.GET("/join/signout", auth.RequireUserMiddleware, func(ctx *gin.Context) {
		Signout(ctx)
		ctx.Redirect(http.StatusFound, "/")
	})
}

func (auth *Auth) authTestRoutes() {

	auth.base.Engine.POST("/join/signin/debug", auth.RequireGuestMiddleware, func(ctx *gin.Context) {

		email := ctx.PostForm("email")
		if email == "" {
			err := errors.New("error: No email found")
			authtemplates.Error(err).Render(ctx, ctx.Writer)
			return
		}
		err := auth.DebugOTP(ctx, email)
		if err != nil {
			authtemplates.Error(err).Render(ctx, ctx.Writer)
			return
		}
		ctx.Redirect(http.StatusFound, "/")
	})

}
