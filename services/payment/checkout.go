package payment

import (
	"fmt"
	"townwatch/base"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

func (payment *Payment) createCheckoutSession(c *paymentmodels.Customer, tierConfig TierConfig) (*stripe.CheckoutSession, *base.ErrorComm) {

	var params *stripe.CheckoutSessionParams
	if !c.StripeCustomerID.Valid {
		params = payment.firstTimerCheckoutParams(c, tierConfig)
	} else {
		params = payment.returningCustomerCheckoutParams(c, tierConfig)
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

func (payment *Payment) firstTimerCheckoutParams(customer *paymentmodels.Customer, tierConfig TierConfig) *stripe.CheckoutSessionParams {
	return &stripe.CheckoutSessionParams{
		// Customer: stripe.String(customerID),
		Mode:          stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		CustomerEmail: stripe.String(customer.Email),
		SuccessURL:    stripe.String(fmt.Sprintf("%s/user/wallet", payment.base.DOMAIN)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(tierConfig.Name),
					},
					Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
						Interval:      stripe.String(tierConfig.Interval),
						IntervalCount: stripe.Int64(1),
					},
					UnitAmount: stripe.Int64(getNewUnitAmount(payment.TierConfigs[customer.TierID], tierConfig)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{"tier": string(tierConfig.TierID)},
	}
}
func (payment *Payment) returningCustomerCheckoutParams(customer *paymentmodels.Customer, tierConfig TierConfig) *stripe.CheckoutSessionParams {
	return &stripe.CheckoutSessionParams{
		Customer: stripe.String(customer.StripeCustomerID.String),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		// CustomerEmail: stripe.String(c.Email),
		// ReturnURL:     stripe.String(payment.base.DOMAIN),
		SuccessURL: stripe.String(fmt.Sprintf("%s/user/wallet", payment.base.DOMAIN)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(tierConfig.Name),
					},
					Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
						Interval:      stripe.String(tierConfig.Interval),
						IntervalCount: stripe.Int64(1),
					},
					UnitAmount: stripe.Int64(getNewUnitAmount(payment.TierConfigs[customer.TierID], tierConfig)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{"tier": string(tierConfig.TierID)},
	}
}

func getNewUnitAmount(currentTierConfig TierConfig, targetTierConfig TierConfig) int64 {
	newCost := targetTierConfig.Amount - currentTierConfig.Amount
	if newCost < 0 {
		newCost = 0
	}
	return newCost
}
