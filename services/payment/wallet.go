package payment

import (
	"fmt"
	"townwatch/base"
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
func (payment *Payment) detachPaymentMethod(paymentMethodID string) *base.ErrorComm {
	_, err := paymentmethod.Detach(paymentMethodID, &stripe.PaymentMethodDetachParams{})
	if err != nil {
		eventId := sentry.CaptureException(err)
		return &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to detach payment (%s)", *eventId),
			DevMsg:  err,
		}
	}
	return nil
}
func (payment *Payment) changeAutoPay(c *paymentmodels.Customer, disable bool) *base.ErrorComm {
	params := &stripe.SubscriptionParams{CancelAtPeriodEnd: stripe.Bool(disable)}
	_, err := subscription.Update(c.StripeSubscriptionID.String, params)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to change autopay (%s)", *eventId),
			DevMsg:  err,
		}
	}
	return nil
}
