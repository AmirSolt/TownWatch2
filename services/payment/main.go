package payment

import (
	"fmt"
	"log"
	"reflect"
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
	// TierConfigs map[paymentmodels.TierID]TierConfig
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

func (payment *Payment) loadStripeProduct() *stripe.Product {
	targetParams := &stripe.ProductParams{
		Name: stripe.String("Premium"),
	}
	var targetProduct *stripe.Product

	params := &stripe.ProductListParams{}
	result := product.List(params)
	for result.Next() {

		productTemp := result.Current().(*stripe.Product)

		// if productTemp.Name == *targetParams.Name {
		if IsASubsetOfB(targetParams, productTemp) {
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
			"tier": string(paymentmodels.Tier1),
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
			"tier": string(paymentmodels.Tier2),
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
			if IsASubsetOfB(targetParams, priceTemp) {
				targetPriceMap[tier] = priceTemp
			}
		}
	}

	for tier, targetPrice := range targetPriceMap {
		if targetPrice == nil {
			result, err := price.New(targetParamsMap[tier])
			if err != nil {
				log.Fatalln("Error: init stripe price load: %w", err)
			}
			targetPriceMap[tier] = result
		}
	}

	return targetPriceMap
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

		// if areStringSlicesEqual(webhook.EnabledEvents, targetParams.EnabledEvents) &&
		// 	webhook.URL == *targetParams.URL &&
		// 	areStringMapsEqual(webhook.Metadata, targetParams.Metadata) {
		// 	targetWebhook = webhook
		if IsASubsetOfB(targetParams, webhook) {
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

func IsASubsetOfB(a, b interface{}) bool {
	// Obtain reflect value objects for both input structs
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// Loop through fields of struct A
	for i := 0; i < aVal.NumField(); i++ {
		// Get field from struct A
		aField := aVal.Field(i)

		// Attempt to find corresponding field in B by name
		bField := bVal.FieldByName(aVal.Type().Field(i).Name)

		// Check if the field exists in B and compare values; return false upon mismatch
		if !bField.IsValid() || aField.Interface() != bField.Interface() {
			return false
		}
	}
	return true
}
