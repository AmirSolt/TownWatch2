package payment

import (
	"context"
	"fmt"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentmethod"
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

func getNewUnitAmount(currentTierConfig TierConfig, targetTierConfig TierConfig) int64 {
	newCost := targetTierConfig.Amount - currentTierConfig.Amount
	if newCost < 0 {
		newCost = 0
	}
	return newCost
}
