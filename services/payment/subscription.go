package payment

import (
	"fmt"
	"townwatch/base"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/subscription"
)

func (payment *Payment) Subscribe(c *paymentmodels.Customer, tier Tier) (*stripe.CheckoutSession, *base.ErrorComm) {
	return payment.createCheckoutSession(c, tier)
}
func (payment *Payment) ChangeSubscriptionTier(c *paymentmodels.Customer, tier Tier) (*stripe.Subscription, *base.ErrorComm) {
	// err := payment.CancelSubscription(c)
	// if err != nil {
	// 	return nil, err
	// }
	// return payment.createCheckoutSession(c, tier)

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			&stripe.SubscriptionItemsParams{
				ID:    stripe.String("{{SUB_ITEM_ID}}"),
				Price: stripe.String(payment.Prices[tier].ID),
			},
		},
	}
	result, err := subscription.Update(c.StripeSubscriptionID.String, params)

	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to change subscription (%s)", *eventId),
			DevMsg:  err,
		}
	}

	return result, nil
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

func (payment *Payment) GetSubscription(subID string) (*stripe.Subscription, *base.ErrorComm) {
	params := &stripe.SubscriptionParams{}
	result, err := subscription.Get(subID, params)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to find subscription (%s)", *eventId),
			DevMsg:  err,
		}
	}

	return result, nil
}
