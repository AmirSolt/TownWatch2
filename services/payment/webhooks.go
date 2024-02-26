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

func (payment *Payment) HandleStripeWebhook(ctx *gin.Context) {
	// ==================================================================
	// The signature check is pulled directly from Stripe and it's not tested
	// req := ctx.Request
	// w := ctx.Writer

	const MaxBodyBytes = int64(65536)
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		ctx.Writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	endpointSecret := payment.base.STRIPE_WEBHOOK_KEY
	event, err := webhook.ConstructEvent(payload, ctx.Request.Header.Get("Stripe-Signature"),
		endpointSecret)
	if err != nil {
		eventId := sentry.CaptureException(err)
		ctx.String(http.StatusBadRequest, fmt.Errorf("error verifying webhook signature. EventID: %s", *eventId).Error())
		return
	}
	// ==================================================================

	if err := payment.handleStripeEvents(event); err != nil {
		eventId := sentry.CaptureException(err)
		ctx.String(http.StatusBadRequest, fmt.Errorf("error handling stripe event. EventID: %s", *eventId).Error())
		return
	}

	ctx.Writer.WriteHeader(http.StatusOK)
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
			return errCust
		}

		errUpd := payment.Queries.UpdateCustomerSubAndTier(context.Background(), paymentmodels.UpdateCustomerSubAndTierParams{
			StripeSubscriptionID: pgtype.Text{String: subsc.ID},
			TierID:               paymentmodels.TierID(tier),
			ID:                   customer.ID,
		})
		if errUpd != nil {
			return errUpd
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
			return errCust
		}

		if customer.StripeSubscriptionID.String == subsc.ID {
			errUpd := payment.Queries.UpdateCustomerSubAndTier(context.Background(), paymentmodels.UpdateCustomerSubAndTierParams{
				StripeSubscriptionID: pgtype.Text{String: "", Valid: false},
				TierID:               paymentmodels.TierIDT0,
				ID:                   customer.ID,
			})
			if errUpd != nil {
				return errUpd
			}
		}
		return nil
	}

	fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	return nil
}
