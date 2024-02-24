package payment

import (
	"fmt"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/subscription"
)

func (payment *Payment) getPaymentMethods(c *paymentmodels.Customer) *customer.PaymentMethodIter {
	params := &stripe.CustomerListPaymentMethodsParams{
		Customer: stripe.String(c.StripeCustomerID.String),
	}
	params.Limit = stripe.Int64(5)
	return customer.ListPaymentMethods(params)
}
func (payment *Payment) detachPaymentMethod(paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, &stripe.PaymentMethodDetachParams{})
	if err != nil {
		eventId := sentry.CaptureException(err)
		return fmt.Errorf("failed to detach payment (%v)", eventId)
	}
	return nil
}
func (payment *Payment) changeAutoPay(c *paymentmodels.Customer, disable bool) error {
	params := &stripe.SubscriptionParams{CancelAtPeriodEnd: stripe.Bool(disable)}
	_, err := subscription.Update(c.StripeSubscriptionID.String, params)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return fmt.Errorf("failed to change autopay (%v)", eventId)
	}
	return nil
}
