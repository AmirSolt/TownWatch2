package payment

import (
	"fmt"
	"townwatch/base"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/subscription"
)

func (payment *Payment) Subscribe(c *paymentmodels.Customer, tierID paymentmodels.TierID) (*stripe.CheckoutSession, *base.ErrorComm) {
	return payment.createCheckoutSession(c, tierID)
}
func (payment *Payment) ChangeSubscriptionTier(c *paymentmodels.Customer, tierID paymentmodels.TierID) (*stripe.CheckoutSession, *base.ErrorComm) {
	err := payment.CancelSubscription(c)
	if err != nil {
		return nil, err
	}
	return payment.createCheckoutSession(c, tierID)
}
func (payment *Payment) CancelSubscription(c *paymentmodels.Customer) *base.ErrorComm {
	_, errSub := subscription.Cancel(c.StripeSubscriptionID.String, &stripe.SubscriptionCancelParams{})
	if errSub != nil {
		eventId := sentry.CaptureException(errSub)
		return &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to cancel subscription (%s)", *eventId),
			DevMsg:  errSub,
		}
	}
	return nil
}
