package base

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

func (base *Base) SendEmail(toEmail, toName, fromName, subject, content string) error {
	payload := EmailPayload{
		ApiKey:      base.EMAIL_CF_WORKER_API_KEY,
		ToEmail:     toEmail,
		ToName:      toName,
		FromName:    fromName,
		Subject:     subject,
		ContentHTML: content,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode payload: %w", err)
	}

	encrypted, err := jwe.Encrypt(payloadBytes, jwe.WithKey(jwa.A128GCM, base.EMAIL_SECRET_KEY))
	if err != nil {
		return fmt.Errorf("failed to encrypt email payload: %w", err)
	}

	resp, err := http.Post(base.EMAIL_CF_WORKER_URL, "application/json", bytes.NewBuffer(encrypted))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("email service returned (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
