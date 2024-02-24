package payment

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"
)

func (payment *Payment) HandleStripeWebhook(ginContext *gin.Context) {
	// ==================================================================
	// The signature check is pulled directly from Stripe and it's not tested
	req := ginContext.Request
	w := ginContext.Writer

	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	endpointSecret := payment.base.STRIPE_WEBHOOK_KEY
	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"),
		endpointSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}
	// ==================================================================

	if err := payment.handleStripeEvents(event); err != nil {
		fmt.Fprintf(os.Stderr, "Error handling event: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (payment *Payment) handleStripeEvents(event stripe.Event) error {
	if event.Type == "customer.subscription.created" {
		cust, err := customer.Get(event.Data.Object["customer"].(string), nil)
		if err != nil {
			return fmt.Errorf("converting raw event to customer object: %w", err)
		}
		subsc, err := subscription.Get(event.Data.Object["subscription"].(string), nil)
		if err != nil {
			return fmt.Errorf("converting raw event to subscription object: %w", err)
		}
		tier := event.Data.Object["metadata"].(string)

		customer, errCust := payment.Queries.GetCustomerByStripeCustomerID(context.Background(), pgtype.Text{String: cust.ID})
		if errCust != nil {
			eventId := sentry.CaptureException(errCust)
			return fmt.Errorf("error eventID: %v", eventId)
		}

		errUpd := payment.Queries.UpdateCustomerSubAndTier(context.Background(), paymentmodels.UpdateCustomerSubAndTierParams{
			StripeSubscriptionID: pgtype.Text{String: subsc.ID},
			TierID:               paymentmodels.TierID(tier),
			ID:                   customer.ID,
		})
		if errUpd != nil {
			eventId := sentry.CaptureException(errUpd)
			return fmt.Errorf("error eventID: %v", eventId)
		}

		return nil
	}

	if event.Type == "customer.subscription.deleted" {
		cust, err := customer.Get(event.Data.Object["customer"].(string), nil)
		if err != nil {
			return fmt.Errorf("converting raw event to customer object: %w", err)
		}
		subsc, err := subscription.Get(event.Data.Object["subscription"].(string), nil)
		if err != nil {
			return fmt.Errorf("converting raw event to subscription object: %w", err)
		}

		customer, errCust := payment.Queries.GetCustomerByStripeCustomerID(context.Background(), pgtype.Text{String: cust.ID})
		if errCust != nil {
			eventId := sentry.CaptureException(errCust)
			return fmt.Errorf("error eventID: %v", eventId)
		}

		if customer.StripeSubscriptionID.String == subsc.ID {
			errUpd := payment.Queries.UpdateCustomerSubAndTier(context.Background(), paymentmodels.UpdateCustomerSubAndTierParams{
				StripeSubscriptionID: pgtype.Text{String: "", Valid: false},
				TierID:               paymentmodels.TierIDT0,
				ID:                   customer.ID,
			})
			if errUpd != nil {
				eventId := sentry.CaptureException(errUpd)
				return fmt.Errorf("error eventID: %v", eventId)
			}
		}
		return nil
	}

	fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	return nil
}
