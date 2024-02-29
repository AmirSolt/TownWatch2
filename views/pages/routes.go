package pages

import (
	"fmt"
	"net/http"
	"townwatch/base"
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

		Page(user, base.IS_PROD, PricingPage(user)).Render(ctx, ctx.Writer)
	})
	base.Engine.GET("/payment/success", auth.RequireUserMiddleware, func(ctx *gin.Context) {
		usertemp, _ := ctx.Get("user")
		user := usertemp.(*authmodels.User)

		Page(user, base.IS_PROD, SuccessPage()).Render(ctx, ctx.Writer)
	})
	base.Engine.GET("/user/wallet", auth.RequireUserMiddleware, func(ctx *gin.Context) {
		usertemp, _ := ctx.Get("user")
		user := usertemp.(*authmodels.User)
		customer, err := payment.Queries.GetCustomerByUserID(ctx, user.ID)
		if err != nil {
			eventId := sentry.CaptureException(err)
			ctx.String(http.StatusBadRequest, fmt.Errorf("failed to find customer (%s)", *eventId).Error())
			return
		}

		subsc, errComm := payment.GetSubscription(customer.StripeSubscriptionID.String)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		Page(user, base.IS_PROD, WalletPage(&customer, subsc, payment.Prices)).Render(ctx, ctx.Writer)
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
