package payment

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"townwatch/base"
	"townwatch/services/auth"
	"townwatch/services/payment/paymentmodels"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
	"github.com/stripe/stripe-go/v76/webhookendpoint"
)

type Payment struct {
	Queries *paymentmodels.Queries
	Prices  map[paymentmodels.Tier]*stripe.Price
	base    *base.Base
	auth    *auth.Auth
}

func LoadPayment(base *base.Base, auth *auth.Auth) *Payment {

	queries := paymentmodels.New(base.Pool)

	payment := Payment{
		Queries: queries,
		base:    base,
		auth:    auth,
		// TierConfigs: loadTierConfigs(),
	}
	payment.loadStripe()
	payment.registerPaymentRoutes()

	return &payment
}

func (payment *Payment) loadStripe() {

	stripe.Key = payment.base.STRIPE_PRIVATE_KEY

	payment.loadStripeWebhook()
	product := payment.loadStripeProduct()
	priceMap := payment.loadStripePrices(product)
	payment.Prices = priceMap

}

func (payment *Payment) loadStripeWebhook() *stripe.WebhookEndpoint {
	targetParams := &stripe.WebhookEndpointParams{
		EnabledEvents: []*string{
			stripe.String("customer.created"),
			// stripe.String("customer.subscription.updated"),
			// stripe.String("customer.subscription.created"),
			// stripe.String("customer.subscription.deleted"),
		},
		URL:      stripe.String(fmt.Sprintf("%s/payment/webhook/events", payment.base.DOMAIN)),
		Metadata: map[string]string{},
	}

	var targetWebhook *stripe.WebhookEndpoint
	params := &stripe.WebhookEndpointListParams{}
	result := webhookendpoint.List(params)
	for result.Next() {

		webhook := result.Current().(*stripe.WebhookEndpoint)
		if webhook.Metadata["params"] == HashStruct(targetParams) {
			targetWebhook = webhook
		} else {
			params := &stripe.WebhookEndpointParams{}
			_, err := webhookendpoint.Del(webhook.ID, params)
			if err != nil {
				log.Fatalln("Error: deleting webhook-endoint: %w", err)
			}
		}
	}

	if targetWebhook == nil {
		var err error

		targetParams.Metadata = map[string]string{
			"params": HashStruct(targetParams),
		}
		targetWebhook, err = webhookendpoint.New(targetParams)
		if err != nil {
			log.Fatalln("Error: init stripe webhook events: %w", err)
		} else {
			log.Fatalln(">>>> a new webhook was created, change the .env webhook_secret: %v", targetWebhook.Secret)
		}
	}

	return targetWebhook
}

func (payment *Payment) loadStripeProduct() *stripe.Product {
	targetParams := &stripe.ProductParams{
		Name: stripe.String("Premium"),
	}

	var targetProduct *stripe.Product
	result := product.List(&stripe.ProductListParams{})
	for result.Next() {

		productTemp := result.Current().(*stripe.Product)

		if productTemp.Metadata["params"] == HashStruct(targetParams) {
			targetProduct = productTemp
		} else {
			params := &stripe.ProductParams{}
			_, err := product.Del(productTemp.ID, params)
			if err != nil {
				log.Fatalln("Error: deleting webhook-endoint: %w", err)
			}
		}
	}

	if targetProduct == nil {
		var err error

		targetParams.Metadata = map[string]string{
			"params": HashStruct(targetParams),
		}
		targetProduct, err = product.New(targetParams)
		if err != nil {
			log.Fatalln("Error: init stripe webhook events: %w", err)
		}
	}

	return targetProduct
}

func (payment *Payment) loadStripePrices(product *stripe.Product) map[paymentmodels.Tier]*stripe.Price {

	targetParamsMonthly := &stripe.PriceParams{
		Nickname: stripe.String("Monthly"),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		Product:  &product.ID,
		Recurring: &stripe.PriceRecurringParams{
			Interval:      stripe.String(string(stripe.PriceRecurringIntervalMonth)),
			IntervalCount: stripe.Int64(1),
		},
		UnitAmount: stripe.Int64(1000),
		Metadata: map[string]string{
			"tier": fmt.Sprintf("%v", paymentmodels.Tier1),
		},
	}
	targetParamsYearly := &stripe.PriceParams{
		Nickname: stripe.String("Yearly"),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		Product:  &product.ID,
		Recurring: &stripe.PriceRecurringParams{
			Interval:      stripe.String(string(stripe.PriceRecurringIntervalYear)),
			IntervalCount: stripe.Int64(1),
		},
		UnitAmount: stripe.Int64(10000),
		Metadata: map[string]string{
			"tier": fmt.Sprintf("%v", paymentmodels.Tier2),
		},
	}

	targetParamsMap := map[paymentmodels.Tier]*stripe.PriceParams{
		paymentmodels.Tier1: targetParamsMonthly,
		paymentmodels.Tier2: targetParamsYearly,
	}

	targetPriceMap := map[paymentmodels.Tier]*stripe.Price{}

	params := &stripe.PriceListParams{}
	result := price.List(params)
	for result.Next() {
		priceTemp := result.Current().(*stripe.Price)
		for tier, targetParams := range targetParamsMap {
			if priceTemp.Metadata["params"] == HashStruct(targetParams) {
				targetPriceMap[tier] = priceTemp
			}
		}
	}

	for tier, targetParams := range targetParamsMap {
		if targetPriceMap[tier] == nil {
			targetParams.Metadata["params"] = HashStruct(targetParams)
			result, err := price.New(targetParams)
			if err != nil {
				log.Fatalln("Error: init stripe price load: %w", err)
			}
			targetPriceMap[tier] = result
		}
	}

	return targetPriceMap
}

func HashStruct(s interface{}) string {
	// Serialize the struct to JSON
	jsonData, err := json.Marshal(s)
	if err != nil {
		log.Fatalf("Error occurred during marshaling. Error: %s", err.Error())
	}

	// Hash the JSON data using SHA-256
	hasher := sha256.New()
	hasher.Write(jsonData)
	hash := hasher.Sum(nil)

	// Encode the hash to a hexadecimal string
	hexStr := hex.EncodeToString(hash)

	// Truncate the hash string to 500 characters if necessary
	// Though with SHA-256, it will never exceed 64 characters
	if len(hexStr) > 500 {
		hexStr = hexStr[:500]
	}

	return hexStr
}
