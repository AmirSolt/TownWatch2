package payment

import (
	"fmt"
	"townwatch/base/basetemplates"
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
			basetemplates.Error(fmt.Errorf("failed to create checkout session (%v)", eventId)).Render(ctx, ctx.Writer)
			return
		}
		checkoutSession, err := payment.Subscribe(&customer, payment.TierConfigs[tierID])
		if err != nil {
			eventId := sentry.CaptureException(err)
			basetemplates.Error(fmt.Errorf("failed to create checkout session (%v)", eventId)).Render(ctx, ctx.Writer)
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
			basetemplates.Error(fmt.Errorf("failed to cancel subscription (%v)", eventId)).Render(ctx, ctx.Writer)
			return
		}
		errCust := payment.CancelSubscription(&customer)
		if errCust != nil {
			eventId := sentry.CaptureException(errCust)
			basetemplates.Error(fmt.Errorf("failed to create checkout session (%v)", eventId)).Render(ctx, ctx.Writer)
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
			basetemplates.Error(fmt.Errorf("failed to create checkout session (%v)", eventId)).Render(ctx, ctx.Writer)
			return
		}
		checkoutSession, err := payment.ChangeSubscriptionTier(&customer, payment.TierConfigs[tierID])
		if err != nil {
			eventId := sentry.CaptureException(err)
			basetemplates.Error(fmt.Errorf("failed to create checkout session (%v)", eventId)).Render(ctx, ctx.Writer)
			return
		}

		ctx.Redirect(302, checkoutSession.URL)
	})

}
