package payment

import (
	"fmt"
	"strconv"
	"townwatch/base"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/subscription"
)

func (payment *Payment) subscribe(c *paymentmodels.Customer, tier paymentmodels.Tier) (*stripe.CheckoutSession, *base.ErrorComm) {
	return payment.createCheckoutSession(c, tier)
}
func (payment *Payment) changeSubscriptionTier(c *paymentmodels.Customer, tier paymentmodels.Tier) (*stripe.Subscription, *base.ErrorComm) {

	subsc, errComm := payment.GetSubscription(c)
	if errComm != nil {
		return nil, errComm
	}
	if subsc == nil {
		err := fmt.Errorf("customer subscription was not found, but requested in changeSubscriptionTier")
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to find subscription (%s)", *eventId),
			DevMsg:  err,
		}
	}

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(subsc.Items.Data[0].ID),
				Price: stripe.String(payment.Prices[tier].ID),
			},
		},
	}
	subsc, err := subscription.Update(subsc.ID, params)

	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to change subscription (%s)", *eventId),
			DevMsg:  err,
		}
	}

	return subsc, nil
}
func (payment *Payment) cancelSubscription(c *paymentmodels.Customer) *base.ErrorComm {

	subsc, errComm := payment.GetSubscription(c)
	if errComm != nil {
		return errComm
	}

	_, errSub := subscription.Cancel(subsc.ID, &stripe.SubscriptionCancelParams{})
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

func (payment *Payment) changeAutoPay(c *paymentmodels.Customer) (*stripe.Subscription, *base.ErrorComm) {

	oldSubsc, errComm := payment.GetSubscription(c)
	if errComm != nil {
		return nil, errComm
	}
	if oldSubsc == nil {
		err := fmt.Errorf("customer subscription was not found, but requested in changeAutoPay")
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to find subscription (%s)", *eventId),
			DevMsg:  err,
		}
	}

	params := &stripe.SubscriptionParams{CancelAtPeriodEnd: stripe.Bool(!oldSubsc.CancelAtPeriodEnd)}
	subsc, err := subscription.Update(oldSubsc.ID, params)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to change autopay (%s)", *eventId),
			DevMsg:  err,
		}
	}
	return subsc, nil
}

func (payment *Payment) GetSubscriptionTier(subsc *stripe.Subscription) (paymentmodels.Tier, *base.ErrorComm) {
	if subsc == nil {
		return paymentmodels.Tier0, nil
	}
	subscTierStr := subsc.Items.Data[0].Price.Metadata["tier"]
	subscTierInt, errTier := strconv.Atoi(subscTierStr)
	if errTier != nil {
		eventId := sentry.CaptureException(errTier)
		return paymentmodels.Tier0, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to find subscription tier (%s)", *eventId),
			DevMsg:  errTier,
		}
	}
	return paymentmodels.Tier(subscTierInt), nil
}

func (payment *Payment) GetSubscription(c *paymentmodels.Customer) (*stripe.Subscription, *base.ErrorComm) {

	params := &stripe.CustomerParams{}
	params.AddExpand("subscriptions")
	result, err := customer.Get(c.StripeCustomerID.String, params)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to find subscription (%s)", *eventId),
			DevMsg:  err,
		}
	}

	var subsc *stripe.Subscription
	for _, subscTemp := range result.Subscriptions.Data {
		subsc = subscTemp
		break
	}
	// if subsc == nil {
	// 	err := fmt.Errorf("customer subscription was not found, stripe subscription search. search params: %+v", params)
	// 	eventId := sentry.CaptureException(err)
	// 	return nil, &base.ErrorComm{
	// 		EventID: eventId,
	// 		UserMsg: fmt.Errorf("failed to find subscription (%s)", *eventId),
	// 		DevMsg:  err,
	// 	}
	// }

	return subsc, nil
}
