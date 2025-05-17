package billing

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func (b *Billing) ApplyCredits(userRepo user.UserRepo, billingRepo BillingRepo, email string, variantID int) error {
	user, err := userRepo.GetUserByEmail(email)
	if err != nil {
		log.Printf("repo.GetUserByEmail failed: %v", err)
		return err
	}

	var (
		credits    int
		creditType string
		reason     string
	)
	switch variantID {
	case b.VariantIDIndividual:
		credits = 1
		creditType = "individual"
		reason = "Individual interview purchase"
	case b.VariantIDPro:
		credits = 10
		creditType = "subscription"
		reason = "Pro subscription monthly credit grant"
	case b.VariantIDPremium:
		credits = 20
		creditType = "subscription"
		reason = "Premium subscription monthly credit grant"
	default:
		log.Printf("ERROR: unknown variantID: %d", variantID)
		return fmt.Errorf("unknown variant ID: %d", variantID)
	}

	if err := userRepo.AddCredits(user.ID, credits, creditType); err != nil {
		log.Printf("repo.AddCredits failed: %v", err)
		return err
	}

	tx := CreditTransaction{
		UserID:     user.ID,
		Amount:     credits,
		CreditType: creditType,
		Reason:     reason,
	}
	if err := billingRepo.LogCreditTransaction(tx); err != nil {
		log.Printf("Warning: credit granted but failed to log transaction: %v", err)
		return err
	}

	return nil
}
