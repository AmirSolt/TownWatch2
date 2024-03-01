package payment

import (
	"fmt"
	"net/http"
	"strconv"
	"townwatch/services/auth/authmodels"
	"townwatch/services/payment/paymentmodels"
	"townwatch/services/payment/paymenttemplates"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func (payment *Payment) registerPaymentRoutes() {
	payment.paymentRoutes()

}

func (payment *Payment) paymentRoutes() {

	payment.base.POST("/subscription/create/:tier", payment.auth.RequireUserMiddleware, func(ctx *gin.Context) {
		tierTmp := ctx.Param("tier")
		tier, err := strconv.Atoi(tierTmp)
		if err != nil {
			eventId := sentry.CaptureException(err)
			ctx.String(http.StatusBadRequest, fmt.Errorf("tier slug is incorrect (%s)", *eventId).Error())
			return
		}

		usertemp, _ := ctx.Get("user")
		user := usertemp.(*authmodels.User)
		customer, err := payment.Queries.GetCustomerByUserID(ctx, user.ID)
		if err != nil {
			eventId := sentry.CaptureException(err)
			ctx.String(http.StatusBadRequest, fmt.Errorf("failed to create checkout session (%s)", *eventId).Error())
			return
		}
		checkoutSession, errComm := payment.subscribe(&customer, paymentmodels.Tier(tier))
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		ctx.Redirect(302, checkoutSession.URL)

	})

	payment.base.POST("/subscription/cancel", payment.auth.RequireUserMiddleware, func(ctx *gin.Context) {
		usertemp, _ := ctx.Get("user")
		user := usertemp.(*authmodels.User)
		customer, err := payment.Queries.GetCustomerByUserID(ctx, user.ID)
		if err != nil {
			eventId := sentry.CaptureException(err)
			ctx.String(http.StatusBadRequest, fmt.Errorf("failed to cancel subscription (%s)", *eventId).Error())
			return
		}
		errComm := payment.cancelSubscription(&customer)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		errRender := paymenttemplates.Tiers(paymentmodels.Tier0, nil, payment.Prices).Render(ctx, ctx.Writer)
		payment.base.HandleRouteRenderError(ctx, errRender)
	})

	payment.base.POST("/subscription/auto/change", payment.auth.RequireUserMiddleware, func(ctx *gin.Context) {

		usertemp, _ := ctx.Get("user")
		user := usertemp.(*authmodels.User)
		customer, err := payment.Queries.GetCustomerByUserID(ctx, user.ID)
		if err != nil {
			eventId := sentry.CaptureException(err)
			ctx.String(http.StatusBadRequest, fmt.Errorf("failed to find customer (%s)", *eventId).Error())
			return
		}

		subsc, errComm := payment.changeAutoPay(&customer)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		tier, errComm := payment.GetSubscriptionTier(subsc)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		errRender := paymenttemplates.Tiers(tier, subsc, payment.Prices).Render(ctx, ctx.Writer)
		payment.base.HandleRouteRenderError(ctx, errRender)
	})

	payment.base.POST("/subscription/change/:tier", payment.auth.RequireUserMiddleware, func(ctx *gin.Context) {

		tierTargetStr := ctx.Param("tier")
		tierTargetInt, err := strconv.Atoi(tierTargetStr)
		if err != nil {
			eventId := sentry.CaptureException(err)
			ctx.String(http.StatusBadRequest, fmt.Errorf("tier slug is incorrect (%s)", *eventId).Error())
			return
		}
		usertemp, _ := ctx.Get("user")
		user := usertemp.(*authmodels.User)
		customer, err := payment.Queries.GetCustomerByUserID(ctx, user.ID)
		if err != nil {
			eventId := sentry.CaptureException(err)
			ctx.String(http.StatusBadRequest, fmt.Errorf("failed to create checkout session (%s)", *eventId).Error())
			return
		}
		subsc, errComm := payment.changeSubscriptionTier(&customer, paymentmodels.Tier(tierTargetInt))
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		tier, errComm := payment.GetSubscriptionTier(subsc)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}
		errRender := paymenttemplates.Tiers(tier, subsc, payment.Prices).Render(ctx, ctx.Writer)
		payment.base.HandleRouteRenderError(ctx, errRender)
	})

	payment.base.POST("/payment/webhook/events", func(ctx *gin.Context) {

		fmt.Println("=================")
		fmt.Printf("\n /payment/webhook/events ctx: %+v \n", ctx)
		fmt.Println("=================")

		payment.HandleStripeWebhook(ctx)
	})

}
