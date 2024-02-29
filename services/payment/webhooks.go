package payment

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"townwatch/services/payment/paymentmodels"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
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
	event, err := webhook.ConstructEvent(payload, ctx.Request.Header.Get("Stripe-Signature"), endpointSecret)
	if err != nil {
		eventId := sentry.CaptureException(err)
		ctx.String(http.StatusBadRequest, fmt.Errorf("error verifying webhook signature. EventID: %s", *eventId).Error())
		return
	}
	// ==================================================================

	if err := payment.handleStripeEvents(ctx, event); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Writer.WriteHeader(http.StatusOK)
}

func (payment *Payment) handleStripeEvents(ctx *gin.Context, event stripe.Event) error {

	if event.Type == "customer.created" {

		stripeCustomer, err := getStripeCustomerFromObj(event.Data.Object)
		if err != nil {
			eventId := sentry.CaptureException(err)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}
		customer, err := payment.Queries.GetCustomerByEmail(ctx, stripeCustomer.Email)
		if err != nil {
			eventId := sentry.CaptureException(err)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}
		errUpd := payment.Queries.UpdateCustomerStripeCustomerID(ctx, paymentmodels.UpdateCustomerStripeCustomerIDParams{
			StripeCustomerID: pgtype.Text{String: stripeCustomer.ID, Valid: true},
			ID:               customer.ID,
		})
		if errUpd != nil {
			eventId := sentry.CaptureException(errUpd)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		return nil
	}
	// =============================================
	if event.Type == "customer.subscription.created" {
		subsc, err := getStripeSubscriptionFromObj(event.Data.Object)
		if err != nil {
			eventId := sentry.CaptureException(err)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		params := &stripe.CustomerParams{}
		stripeCustomer, err := customer.Get(subsc.Customer.ID, params)
		if err != nil {
			eventId := sentry.CaptureException(err)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		fmt.Println("==================================")
		fmt.Printf("\n>> stripeCustomer: %+v \n\n", stripeCustomer)
		fmt.Println("==================================")

		customer, errCust := payment.Queries.GetCustomerByEmail(ctx, stripeCustomer.Email)
		if errCust != nil {
			eventId := sentry.CaptureException(errCust)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		tierStr := subsc.Items.Data[0].Price.Metadata["tier"]
		tier, err := strconv.Atoi(tierStr)
		if err != nil {
			eventId := sentry.CaptureException(err)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		errUpd := payment.Queries.UpdateCustomerSubAndTier(ctx, paymentmodels.UpdateCustomerSubAndTierParams{
			StripeSubscriptionID: pgtype.Text{String: subsc.ID, Valid: true},
			Tier:                 int32(paymentmodels.Tier(tier)),
			ID:                   customer.ID,
		})
		if errUpd != nil {
			eventId := sentry.CaptureException(errUpd)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		return nil
	}

	// =============================================

	if event.Type == "customer.subscription.updated" {
		subsc, err := getStripeSubscriptionFromObj(event.Data.Object)
		if err != nil {
			eventId := sentry.CaptureException(err)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		params := &stripe.CustomerParams{}
		stripeCustomer, err := customer.Get(subsc.Customer.ID, params)
		if err != nil {
			eventId := sentry.CaptureException(err)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		fmt.Println("==================================")
		fmt.Printf("\n>> stripeCustomer: %+v \n\n", stripeCustomer)
		fmt.Println("==================================")

		customer, errCust := payment.Queries.GetCustomerByEmail(ctx, stripeCustomer.Email)
		if errCust != nil {
			eventId := sentry.CaptureException(errCust)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		tierStr := subsc.Items.Data[0].Price.Metadata["tier"]
		tier, err := strconv.Atoi(tierStr)
		if err != nil {
			eventId := sentry.CaptureException(err)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		errUpd := payment.Queries.UpdateCustomerSubAndTier(ctx, paymentmodels.UpdateCustomerSubAndTierParams{
			StripeSubscriptionID: pgtype.Text{String: subsc.ID, Valid: true},
			Tier:                 int32(paymentmodels.Tier(tier)),
			ID:                   customer.ID,
		})
		if errUpd != nil {
			eventId := sentry.CaptureException(errUpd)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		return nil
	}

	// =============================================
	if event.Type == "customer.subscription.deleted" {
		subsc, err := getStripeSubscriptionFromObj(event.Data.Object)
		if err != nil {
			eventId := sentry.CaptureException(err)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		customer, errCust := payment.Queries.GetCustomerByStripeCustomerID(ctx, pgtype.Text{String: subsc.Customer.ID, Valid: true})
		if errCust != nil {
			eventId := sentry.CaptureException(errCust)
			return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
		}

		if customer.StripeSubscriptionID.String == subsc.ID {
			errUpd := payment.Queries.UpdateCustomerSubAndTier(ctx, paymentmodels.UpdateCustomerSubAndTierParams{
				StripeSubscriptionID: pgtype.Text{String: "", Valid: false},
				Tier:                 int32(paymentmodels.Tier0),
				ID:                   customer.ID,
			})
			if errUpd != nil {
				eventId := sentry.CaptureException(errUpd)
				return fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
			}
		}
		return nil
	}

	fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	return nil
}

func getStripeCustomerFromObj(object map[string]interface{}) (*stripe.Customer, error) {
	jsonCustomer, err := json.Marshal(object)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
	}
	var stripeCustomer *stripe.Customer
	err = json.Unmarshal(jsonCustomer, &stripeCustomer)
	if stripeCustomer == nil || err != nil {
		eventId := sentry.CaptureException(err)
		return nil, fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
	}
	return stripeCustomer, nil
}

func getStripeCheckoutSessionFromObj(object map[string]interface{}) (*stripe.CheckoutSession, error) {
	jsonCustomer, err := json.Marshal(object)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
	}
	var stripeStruct *stripe.CheckoutSession
	err = json.Unmarshal(jsonCustomer, &stripeStruct)
	if stripeStruct == nil || err != nil {
		eventId := sentry.CaptureException(err)
		return nil, fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
	}
	return stripeStruct, nil
}

func getStripeSubscriptionFromObj(object map[string]interface{}) (*stripe.Subscription, error) {
	jsonCustomer, err := json.Marshal(object)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
	}
	var stripeStruct *stripe.Subscription
	err = json.Unmarshal(jsonCustomer, &stripeStruct)
	if stripeStruct == nil || err != nil {
		eventId := sentry.CaptureException(err)
		return nil, fmt.Errorf("error handling stripe event. EventID: %s", *eventId)
	}
	return stripeStruct, nil
}
