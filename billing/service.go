package billing

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/michaelboegner/interviewer/user"
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

func (b *Billing) UpdateSubscription(repo *user.UserRepo, payload BillingWebhookPayload) error {
	user, err := repo.GetUserByCustomerID(payload.Data.Attributes.CustomerID)
	if err != nil {
		log.Printf("repo.GetUserByCustomerID failed: %v", err)
		return err
	}

	user.SubscriptionTier = payload.Data.Attributes.VariantID
	user.BillingSubscriptionID = payload.Data.Attributes.SubscriptionID
	user.BillingStatus = payload.Data.Attributes.Status
	user.SubscriptionStartDate = time.Now().UTC()

	err = repo.UpdateBillingInfo(user)
	if err != nil {
		log.Printf("repo.UpdateBillingInfo failed: %v", err)
		return err
	}

	return nil
}

// func CancelSubscription(repo UserRepo, payload billing.BillingWebhookPayload) error {
// 	user, err := repo.GetUserByCustomerID(payload.Data.Attributes.CustomerID)
// 	if err != nil {
// 		log.Printf("repo.GetUserByCustomerID failed: %v", err)
// 		return err
// 	}

// 	user.BillingStatus = "cancelled"

// 	err = repo.UpdateBillingInfo(user)
// 	if err != nil {
// 		log.Printf("repo.UpdateBillingInfo failed: %v", err)
// 		return err
// 	}

// 	return nil
// }
