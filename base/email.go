package base

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
)

type EmailPayload struct {
	ApiKey      string `json:"apiKey"`
	ToEmail     string `json:"toEmail"`
	ToName      string `json:"toName"`
	FromName    string `json:"fromName"`
	Subject     string `json:"subject"`
	ContentHTML string `json:"contentHTML"`
}

func (base *Base) SendEmail(toEmail, toName, fromName, subject, content string) *ErrorComm {
	payload := EmailPayload{
		ApiKey:      base.EMAIL_CF_WORKER_API_KEY,
		ToEmail:     toEmail,
		ToName:      toName,
		FromName:    fromName,
		Subject:     subject,
		ContentHTML: content,
	}
	fmt.Println("==================================")
	fmt.Printf("\n>> Email Payload: %+v \n\n", payload)
	fmt.Println("==================================")

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return &ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to send email (%s)", *eventId),
			DevMsg:  err,
		}
	}

	encrypted, err := jwe.Encrypt(payloadBytes, jwe.WithKey(jwa.A128GCMKW, []byte(base.EMAIL_SECRET_KEY)))
	if err != nil {
		eventId := sentry.CaptureException(err)
		return &ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to send email (%s)", *eventId),
			DevMsg:  err,
		}
	}

	resp, err := http.Post(base.EMAIL_CF_WORKER_URL, "application/json", bytes.NewBuffer(encrypted))
	if err != nil {
		eventId := sentry.CaptureException(err)
		return &ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to send email (%s)", *eventId),
			DevMsg:  err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		eventId := sentry.CaptureException(fmt.Errorf("email failed response. response: %v | ", string(bodyBytes)))
		return &ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("failed to send email (%s)", *eventId),
			DevMsg:  err,
		}
	}

	return nil
}
