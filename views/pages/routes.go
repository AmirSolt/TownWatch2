package pages

import (
	"townwatch/base"

	"github.com/gin-gonic/gin"
)

func RegisterPagesRoutes(base *base.Base) {

	// TODO: middleware
	base.Engine.GET("/join", func(ctx *gin.Context) {
		JoinPage().Render(ctx, ctx.Writer)
	})
	// TODO: middleware
	base.Engine.GET("/join/verif", func(ctx *gin.Context) {
		VerifyPage().Render(ctx, ctx.Writer)
	})

	base.Engine.GET("/", func(ctx *gin.Context) {
		IndexPage(base.IS_PROD).Render(ctx, ctx.Writer)
	})

}
