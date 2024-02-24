package payment

import (
	"fmt"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/subscription"
)

func (payment *Payment) Subscribe(c *paymentmodels.Customer, tierConfig TierConfig) (*stripe.CheckoutSession, error) {
	return payment.createCheckoutSession(c, tierConfig)
}
func (payment *Payment) ChangeSubscriptionTier(c *paymentmodels.Customer, tierConfig TierConfig) (*stripe.CheckoutSession, error) {
	err := payment.CancelSubscription(c)
	if err != nil {
		return nil, err
	}
	return payment.createCheckoutSession(c, tierConfig)
}
func (payment *Payment) CancelSubscription(c *paymentmodels.Customer) error {
	_, errSub := subscription.Cancel(c.StripeSubscriptionID.String, &stripe.SubscriptionCancelParams{})
	if errSub != nil {
		eventId := sentry.CaptureException(errSub)
		return fmt.Errorf("failed to cancel subscription (%v)", eventId)
	}
	return nil
}
