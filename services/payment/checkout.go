package payment

import (
	"fmt"
	"townwatch/base"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

func (payment *Payment) createCheckoutSession(c *paymentmodels.Customer, tier paymentmodels.Tier) (*stripe.CheckoutSession, *base.ErrorComm) {

	var customerID *string
	var customerEmail *string
	if !c.StripeCustomerID.Valid {
		customerID = nil
		customerEmail = &c.Email
	} else {
		customerID = &c.StripeCustomerID.String
		customerEmail = nil
	}

	var params *stripe.CheckoutSessionParams = &stripe.CheckoutSessionParams{
		Customer:      customerID,
		Mode:          stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		CustomerEmail: customerEmail,
		SuccessURL:    stripe.String(fmt.Sprintf("%s/payment/success", payment.base.DOMAIN)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    &payment.Prices[tier].ID,
				Quantity: stripe.Int64(1),
			},
		},
	}

	result, err := session.New(params)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("checkout session creation failed (%s)", *eventId),
			DevMsg:  err,
		}
	}

	return result, nil
}
