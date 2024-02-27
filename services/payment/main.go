package payment

import (
	"fmt"
	"log"
	"townwatch/base"
	"townwatch/services/auth"
	"townwatch/services/payment/paymentmodels"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhookendpoint"
)

type Payment struct {
	Queries     *paymentmodels.Queries
	base        *base.Base
	auth        *auth.Auth
	TierConfigs map[paymentmodels.TierID]TierConfig
}

type TierConfig struct {
	TierID   paymentmodels.TierID
	Name     string
	Symbol   string
	Interval string
	Amount   int64
	Level    int64
}

func LoadPayment(base *base.Base, auth *auth.Auth) *Payment {

	queries := paymentmodels.New(base.Pool)

	payment := Payment{
		Queries:     queries,
		base:        base,
		auth:        auth,
		TierConfigs: loadTierConfigs(),
	}
	payment.loadStripe()
	payment.registerPaymentRoutes()

	return &payment
}

func loadTierConfigs() map[paymentmodels.TierID]TierConfig {
	m := make(map[paymentmodels.TierID]TierConfig)
	m[paymentmodels.TierIDT0] = TierConfig{
		TierID:   paymentmodels.TierIDT0,
		Name:     "Free",
		Interval: "never",
		Symbol:   "$",
		Amount:   0,
	}
	m[paymentmodels.TierIDT1] = TierConfig{
		TierID:   paymentmodels.TierIDT1,
		Name:     "Monthly",
		Interval: "month",
		Symbol:   "$",
		Amount:   1000,
	}
	m[paymentmodels.TierIDT2] = TierConfig{
		TierID:   paymentmodels.TierIDT2,
		Name:     "Yearly",
		Interval: "year",
		Symbol:   "$",
		Amount:   10000,
	}
	return m
}

func (payment *Payment) loadStripe() {

	stripe.Key = payment.base.STRIPE_PRIVATE_KEY

	product := payment.loadStripeProduct()
	payment.loadStripePrices(product)
	payment.loadStripeWebhook()

}

func (payment *Payment) loadStripeProduct() *stripe.Product {
	// same as webhook enpoint
	return nil
}

func (payment *Payment) loadStripePrices(product *stripe.Product) []*stripe.Price {

}

func (payment *Payment) loadStripeWebhook() *stripe.WebhookEndpoint {
	targetParams := &stripe.WebhookEndpointParams{
		EnabledEvents: []*string{
			stripe.String("customer.created"),
			stripe.String("customer.subscription.created"),
			stripe.String("customer.subscription.deleted"),
		},
		URL:      stripe.String(fmt.Sprintf("%s/payment/webhook/events", payment.base.DOMAIN)),
		Metadata: map[string]string{},
	}
	var targetWebhook *stripe.WebhookEndpoint

	params := &stripe.WebhookEndpointListParams{}
	result := webhookendpoint.List(params)
	for result.Next() {

		webhook := result.Current().(*stripe.WebhookEndpoint)

		if areStringSlicesEqual(webhook.EnabledEvents, targetParams.EnabledEvents) &&
			webhook.URL == *targetParams.URL &&
			areStringMapsEqual(webhook.Metadata, targetParams.Metadata) {
			targetWebhook = webhook
		} else {
			params := &stripe.WebhookEndpointParams{}
			_, err := webhookendpoint.Del(payment.base.STRIPE_WEBHOOK_KEY, params)
			if err != nil {
				log.Fatalln("Error: deleting webhook-endoint: %w", err)
			}
		}
	}

	if targetWebhook == nil {
		var err error
		targetWebhook, err = webhookendpoint.New(targetParams)
		if err != nil {
			log.Fatalln("Error: init stripe webhook events: %w", err)
		}
	}

	return targetWebhook
}

func areStringSlicesEqual(strs []string, ptrs []*string) bool {
	if len(strs) != len(ptrs) {
		return false
	}
	for i, str := range strs {
		if ptrs[i] == nil || str != *ptrs[i] {
			return false
		}
	}
	return true
}

func areStringMapsEqual(map1, map2 map[string]string) bool {
	if len(map1) != len(map2) {
		return false
	}
	for key, value := range map1 {
		if val2, ok := map2[key]; !ok || value != val2 {
			return false
		}
	}
	return true
}
