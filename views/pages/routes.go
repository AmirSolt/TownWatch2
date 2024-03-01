package pages

import (
	"fmt"
	"net/http"
	"townwatch/base"
	"townwatch/services/auth"
	"townwatch/services/auth/authmodels"
	"townwatch/services/payment"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func RegisterPagesRoutes(base *base.Base, auth *auth.Auth, payment *payment.Payment) {

	base.Engine.GET("/join", auth.RequireGuestMiddleware, func(ctx *gin.Context) {
		errRender := PageNoLayout(JoinPage()).Render(ctx, ctx.Writer)
		base.HandleRouteRenderError(ctx, errRender)
	})
	base.Engine.GET("/join/verif", auth.RequireGuestMiddleware, func(ctx *gin.Context) {
		errRender := PageNoLayout(VerifyPage()).Render(ctx, ctx.Writer)
		base.HandleRouteRenderError(ctx, errRender)
	})

	// ================================
	base.Engine.GET("/pricing", auth.OptionalUserMiddleware, func(ctx *gin.Context) {
		usertemp, exists := ctx.Get("user")
		var user *authmodels.User
		if exists {
			user = usertemp.(*authmodels.User)
		}

		var tier paymentmodels.Tier = paymentmodels.Tier0
		if user != nil {

			customer, err := payment.Queries.GetCustomerByUserID(ctx, user.ID)
			if err != nil {
				eventId := sentry.CaptureException(err)
				ctx.String(http.StatusBadRequest, fmt.Errorf("failed to find customer (%s)", *eventId).Error())
				return
			}

			subsc, errComm := payment.GetSubscription(&customer)
			if errComm != nil {
				ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
				return
			}

			tierTemp, errCommTier := payment.GetSubscriptionTier(subsc)
			tier = tierTemp
			if errCommTier != nil {
				ctx.String(http.StatusBadRequest, errCommTier.UserMsg.Error())
				return
			}
		}

		errRender := Page(user, base.IS_PROD, PricingPage(user, tier)).Render(ctx, ctx.Writer)
		base.HandleRouteRenderError(ctx, errRender)
	})
	base.Engine.GET("/payment/success", auth.RequireUserMiddleware, func(ctx *gin.Context) {
		usertemp, _ := ctx.Get("user")
		user := usertemp.(*authmodels.User)

		errRender := Page(user, base.IS_PROD, SuccessPage()).Render(ctx, ctx.Writer)
		base.HandleRouteRenderError(ctx, errRender)
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

		subsc, errComm := payment.GetSubscription(&customer)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		tier, errComm := payment.GetSubscriptionTier(subsc)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		errRender := Page(user, base.IS_PROD, WalletPage(tier, subsc, payment.Prices)).Render(ctx, ctx.Writer)
		base.HandleRouteRenderError(ctx, errRender)
	})
	// ================================

	base.Engine.GET("/", auth.OptionalUserMiddleware, func(ctx *gin.Context) {
		usertemp, exists := ctx.Get("user")
		var user *authmodels.User
		if exists {
			user = usertemp.(*authmodels.User)
		}
		errRender := Page(user, base.IS_PROD, IndexPage()).Render(ctx, ctx.Writer)
		base.HandleRouteRenderError(ctx, errRender)
	})

}
