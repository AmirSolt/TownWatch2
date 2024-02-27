package payment

import (
	"fmt"
	"net/http"
	"townwatch/services/auth/authmodels"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func (payment *Payment) registerPaymentRoutes() {
	payment.paymentRoutes()

}

func (payment *Payment) paymentRoutes() {

	payment.base.GET("/subscription/create/:tierID", payment.auth.RequireUserMiddleware, func(ctx *gin.Context) {
		tierIDTemp := ctx.Param("tierID")
		tierID := paymentmodels.TierID(tierIDTemp)
		usertemp, _ := ctx.Get("user")
		user := usertemp.(*authmodels.User)
		customer, err := payment.Queries.GetCustomerByUserID(ctx, user.ID)
		if err != nil {
			eventId := sentry.CaptureException(err)
			ctx.String(http.StatusBadRequest, fmt.Errorf("failed to create checkout session (%s)", *eventId).Error())
			return
		}
		checkoutSession, errComm := payment.Subscribe(&customer, tierID)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		ctx.Redirect(302, checkoutSession.URL)

	})

	payment.base.GET("/subscription/cancel/:tierID", payment.auth.RequireUserMiddleware, func(ctx *gin.Context) {
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

	payment.base.GET("/subscription/change/:tierID", payment.auth.RequireUserMiddleware, func(ctx *gin.Context) {
		tierIDTemp := ctx.Param("tierID")
		tierID := paymentmodels.TierID(tierIDTemp)
		usertemp, _ := ctx.Get("user")
		user := usertemp.(*authmodels.User)
		customer, err := payment.Queries.GetCustomerByUserID(ctx, user.ID)
		if err != nil {
			eventId := sentry.CaptureException(err)
			ctx.String(http.StatusBadRequest, fmt.Errorf("failed to create checkout session (%s)", *eventId).Error())
			return
		}
		checkoutSession, errComm := payment.ChangeSubscriptionTier(&customer, tierID)
		if errComm != nil {
			ctx.String(http.StatusBadRequest, errComm.UserMsg.Error())
			return
		}

		ctx.Redirect(302, checkoutSession.URL)
	})

	payment.base.POST("/payment/webhook/events", func(ctx *gin.Context) {

		fmt.Println("=================")
		fmt.Printf("\n /payment/webhook/events ctx: %+v \n", ctx)
		fmt.Println("=================")

		payment.HandleStripeWebhook(ctx)
	})

}
