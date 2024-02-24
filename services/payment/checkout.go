package payment

import (
	"context"
	"fmt"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

func (payment *Payment) createCheckoutSession(c *paymentmodels.Customer, tierConfig TierConfig) (*stripe.CheckoutSession, error) {

	var customerID *string = nil
	if c.StripeCustomerID.Valid {
		customerID = &c.StripeCustomerID.String
	}

	params := &stripe.CheckoutSessionParams{
		Customer:      customerID,
		Mode:          stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		CustomerEmail: stripe.String(c.Email),
		ReturnURL:     stripe.String(payment.base.DOMAIN),
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
					UnitAmount: stripe.Int64(getNewUnitAmount(payment.TierConfigs[c.TierID], tierConfig)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{"tier": string(tierConfig.TierID)},
	}

	result, err := session.New(params)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, fmt.Errorf("checkout session creation failed (%v)", eventId)
	}

	if customerID == nil {
		err := payment.Queries.UpdateCustomerStripeCustomerID(context.Background(), paymentmodels.UpdateCustomerStripeCustomerIDParams{
			StripeCustomerID: pgtype.Text{String: result.Customer.ID},
			ID:               c.ID,
		})
		if err != nil {
			eventId := sentry.CaptureException(err)
			return nil, fmt.Errorf("checkout session creation failed (%v)", eventId)
		}
	}

	return result, nil
}

func getNewUnitAmount(currentTierConfig TierConfig, targetTierConfig TierConfig) int64 {
	newCost := targetTierConfig.Amount - currentTierConfig.Amount
	if newCost < 0 {
		newCost = 0
	}
	return newCost
}
