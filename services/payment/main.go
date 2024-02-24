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
	Interval string
	Amount   int64
	Level    int64
}

func LoadAuth(base *base.Base, auth *auth.Auth) *Payment {

	queries := paymentmodels.New(base.Conn)

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

func (payment *Payment) loadStripe() {
	// stripe key
	stripe.Key = payment.base.STRIPE_PRIVATE_KEY

	// webhook setup
	params := &stripe.WebhookEndpointParams{
		EnabledEvents: []*string{
			// stripe.String("customer.subscription.updated"),
			stripe.String("customer.subscription.created"),
			stripe.String("customer.subscription.deleted"),
			// stripe.String("customer.subscription.resumed"),
			// stripe.String("customer.subscription.paused"),
			// stripe.String("payment_method.attached"),
			// stripe.String("payment_method.detached"),
		},
		URL: stripe.String(fmt.Sprintf("%s/payment/webhook/events", payment.base.DOMAIN)),
	}
	_, err := webhookendpoint.New(params)
	if err != nil {
		log.Fatalln("Error: init stripe webhook events: %w", err)
	}
}

func loadTierConfigs() map[paymentmodels.TierID]TierConfig {
	m := make(map[paymentmodels.TierID]TierConfig)
	m[paymentmodels.TierIDT0] = TierConfig{
		TierID:   paymentmodels.TierIDT0,
		Name:     "Free",
		Interval: "never",
		Amount:   0,
	}
	m[paymentmodels.TierIDT1] = TierConfig{
		TierID:   paymentmodels.TierIDT1,
		Name:     "Monthly",
		Interval: "month",
		Amount:   1000,
	}
	m[paymentmodels.TierIDT2] = TierConfig{
		TierID:   paymentmodels.TierIDT2,
		Name:     "Yearly",
		Interval: "year",
		Amount:   10000,
	}
	return m
}