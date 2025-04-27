package billing

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

func (b *Billing) CreateCheckoutSession(userEmail string, variantID int) (string, error) {
	payload := CheckoutPayload{
		VariantID: variantID,
		Email:     userEmail,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("json.Marshal failed: %v", err)
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.lemonsqueezy.com/v1/checkouts", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("http.NewRequest failed: %v", err)
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+b.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.api+json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("client.Do failed: %v", err)
		return "", err
	}
	defer res.Body.Close()

	var result CheckoutResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Printf("NewDecoder failed: %v", err)
		return "", err
	}

	return result.Data.Attributes.URL, nil
}

func (b *Billing) ServiceWebhook(webhookPayload BillingWebhookPayload) error {

	eventType := webhookPayload.Meta.EventName
	switch eventType {
	case "subscription_created", "subscription_updated":
		//user service update
	case "subscription_cancelled":
		// user service cancellation
	default:
		log.Printf("Unhandled event type: %s", eventType)
		return errors.New("Unhandled event type")
	}

	return nil
}
