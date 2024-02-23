package pages

import (
	"fmt"
	"townwatch/base"
	"townwatch/services/auth"
	"townwatch/services/auth/authmodels"

	"github.com/gin-gonic/gin"
)

func RegisterPagesRoutes(base *base.Base, auth *auth.Auth) {

	base.Engine.GET("/join", auth.RequireGuestMiddleware, func(ctx *gin.Context) {
		JoinPage().Render(ctx, ctx.Writer)
	})
	base.Engine.GET("/join/verif", auth.RequireGuestMiddleware, func(ctx *gin.Context) {
		VerifyPage().Render(ctx, ctx.Writer)
	})

	base.Engine.GET("/", auth.OptionalUserMiddleware, func(ctx *gin.Context) {
		usertemp, exists := ctx.Get("user")
		var user *authmodels.User
		if exists {
			user = usertemp.(*authmodels.User)
		}

		fmt.Println(".>>>>>>>>>>", user)

		IndexPage(user, base.IS_PROD).Render(ctx, ctx.Writer)
	})

}
