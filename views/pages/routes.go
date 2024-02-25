package pages

import (
	"fmt"
	"townwatch/base"
	"townwatch/base/basetemplates"
	"townwatch/services/auth"
	"townwatch/services/auth/authmodels"
	"townwatch/services/payment"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func RegisterPagesRoutes(base *base.Base, auth *auth.Auth, payment *payment.Payment) {

	base.Engine.GET("/join", auth.RequireGuestMiddleware, func(ctx *gin.Context) {
		PageNoLayout(JoinPage()).Render(ctx, ctx.Writer)
	})
	base.Engine.GET("/join/verif", auth.RequireGuestMiddleware, func(ctx *gin.Context) {
		PageNoLayout(VerifyPage()).Render(ctx, ctx.Writer)
	})

	// ================================
	base.Engine.GET("/pricing", auth.OptionalUserMiddleware, func(ctx *gin.Context) {
		usertemp, exists := ctx.Get("user")
		var user *authmodels.User
		if exists {
			user = usertemp.(*authmodels.User)
		}

		Page(user, base.IS_PROD, PricingPage()).Render(ctx, ctx.Writer)
	})
	base.Engine.GET("/user/wallet", auth.RequireUserMiddleware, func(ctx *gin.Context) {
		usertemp, _ := ctx.Get("user")
		user := usertemp.(*authmodels.User)
		customer, err := payment.Queries.GetCustomerByUserID(ctx, user.ID)
		if err != nil {
			eventId := sentry.CaptureException(err)
			basetemplates.Error(fmt.Errorf("failed to cancel subscription (%v)", eventId)).Render(ctx, ctx.Writer)
		}

		Page(user, base.IS_PROD, WalletPage(&customer, payment.TierConfigs)).Render(ctx, ctx.Writer)
	})
	// ================================

	base.Engine.GET("/", auth.OptionalUserMiddleware, func(ctx *gin.Context) {
		usertemp, exists := ctx.Get("user")
		var user *authmodels.User
		if exists {
			user = usertemp.(*authmodels.User)
		}
		Page(user, base.IS_PROD, IndexPage()).Render(ctx, ctx.Writer)
	})

}
