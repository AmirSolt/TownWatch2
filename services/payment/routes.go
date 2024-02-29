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

	payment.base.GET("/subscription/create/:tier", payment.auth.RequireUserMiddleware, func(ctx *gin.Context) {
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
		checkoutSession, errComm := payment.Subscribe(&customer, paymentmodels.Tier(tier))
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		ctx.Redirect(302, checkoutSession.URL)

	})

	payment.base.GET("/subscription/cancel/:tier", payment.auth.RequireUserMiddleware, func(ctx *gin.Context) {
		usertemp, _ := ctx.Get("user")
		user := usertemp.(*authmodels.User)
		customer, err := payment.Queries.GetCustomerByUserID(ctx, user.ID)
		if err != nil {
			eventId := sentry.CaptureException(err)
			ctx.String(http.StatusBadRequest, fmt.Errorf("failed to cancel subscription (%s)", *eventId).Error())
			return
		}
		errComm := payment.CancelSubscription(&customer)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		ctx.Redirect(302, "/user/wallet")
	})

	payment.base.POST("/subscription/change/:tier", payment.auth.RequireUserMiddleware, func(ctx *gin.Context) {

		tierTemp := ctx.Param("tier")
		tier, err := strconv.Atoi(tierTemp)
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
		subsc, errComm := payment.ChangeSubscriptionTier(&customer, paymentmodels.Tier(tier))
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		paymenttemplates.Tiers(&customer, subsc, payment.Prices).Render(ctx, ctx.Writer)

		// paymenttemplates.WalletTier(&customer, subsc, Tier(tier), payment.Prices[Tier(tier)], payment.Prices).Render(ctx, ctx.Writer)

		// ctx.Redirect(http.StatusPermanentRedirect, "/user/wallet")
	})

	payment.base.POST("/payment/webhook/events", func(ctx *gin.Context) {

		fmt.Println("=================")
		fmt.Printf("\n /payment/webhook/events ctx: %+v \n", ctx)
		fmt.Println("=================")

		payment.HandleStripeWebhook(ctx)
	})

}
