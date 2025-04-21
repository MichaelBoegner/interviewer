package customerio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Mailer struct {
	APIKey  string
	APIPass string
	BaseURL string
}

func New(apiKey string) *Mailer {
	return &Mailer{
		APIKey:  os.Getenv("CIO_KEY"),
		APIPass: os.Getenv("CIO_PASS"),
		BaseURL: "https://track.customer.io/api/v1",
	}
}

func (m *Mailer) SendPasswordReset(email, resetURL string) error {
	payload := map[string]any{
		"email":     email,
		"reset_url": resetURL,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Marshal failed: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/customers/%s/events/password_reset_requested", m.BaseURL, email), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(m.APIKey, m.APIPass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Customer.io error: %s", resp.Status)
	}

	return nil
}
