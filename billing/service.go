package billing

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/michaelboegner/interviewer/user"
)

func (b *Billing) RequestCheckoutSession(userEmail string, variantID int) (string, error) {
	payload := CheckoutPayload{
		Data: CheckoutData{
			Type: "checkouts",
			Attributes: CheckoutAttributes{
				CheckoutData: CheckoutCustomerInfo{
					Email: userEmail,
				},
			},
			Relationships: CheckoutRelationships{
				Store: Relationship{
					Data: RelationshipData{
						Type: "stores",
						ID:   os.Getenv("LEMON_STORE_ID"),
					},
				},
				Variant: Relationship{
					Data: RelationshipData{
						Type: "variants",
						ID:   strconv.Itoa(variantID),
					},
				},
			},
		},
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

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(res.Body)
		log.Printf("LemonSqueezy returned error: %s", string(bodyBytes))
		return "", fmt.Errorf("LemonSqueezy API error: %s", res.Status)
	}

	var result CheckoutResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Printf("NewDecoder failed: %v", err)
		return "", err
	}

	return result.Data.Attributes.URL, nil
}

func (b *Billing) RequestDeleteSubscription(subscriptionID string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("DELETE", "https://api.lemonsqueezy.com/v1/subscriptions/"+subscriptionID, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+b.APIKey)
	req.Header.Set("Accept", "application/vnd.api+json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("cancel failed: %s", string(bodyBytes))
	}

	return nil
}

func (b *Billing) RequestResumeSubscription(subscriptionID string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "subscriptions",
			"id":   subscriptionID,
			"attributes": map[string]bool{
				"cancelled": false,
			},
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("PATCH", "https://api.lemonsqueezy.com/v1/subscriptions/"+subscriptionID, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+b.APIKey)
	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("Content-Type", "application/vnd.api+json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("cancel failed: %s", string(bodyBytes))
	}

	return nil
}

func (b *Billing) RequestUpdateSubscriptionVariant(subscriptionID string, newVariantID int) error {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "subscriptions",
			"id":   subscriptionID,
			"attributes": map[string]interface{}{
				"variant_id": newVariantID,
			},
		},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PATCH", "https://api.lemonsqueezy.com/v1/subscriptions/"+subscriptionID, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+b.APIKey)
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/vnd.api+json")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("patch failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("lemon error: %s", string(bodyBytes))
	}
	return nil
}

func (b *Billing) VerifyBillingSignature(signature string, body []byte, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
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

func (b *Billing) DeductCredits(userRepo user.UserRepo, billingRepo BillingRepo, orderAttrs OrderAttributes) error {
	user, err := userRepo.GetUserByEmail(orderAttrs.UserEmail)
	if err != nil {
		log.Printf("repo.GetUserByEmail failed: %v", err)
		return err
	}

	var (
		credits    int
		creditType string
		reason     string
	)

	variantID := orderAttrs.FirstOrderItem.VariantID
	switch variantID {
	case b.VariantIDIndividual:
		credits = 1
		creditType = "individual"
		reason = "Individual interview refund"
	case b.VariantIDPro:
		credits = 10
		creditType = "subscription"
		reason = "Pro subscription monthly credit refund"
	case b.VariantIDPremium:
		credits = 20
		creditType = "subscription"
		reason = "Premium subscription monthly credit refund"
	default:
		log.Printf("ERROR: unknown variantID: %d", variantID)
		return fmt.Errorf("unknown variant ID: %d", variantID)
	}

	if err := userRepo.AddCredits(user.ID, -credits, creditType); err != nil {
		log.Printf("repo.DeductCredits failed: %v", err)
		return err
	}

	tx := CreditTransaction{
		UserID:     user.ID,
		Amount:     -credits,
		CreditType: creditType,
		Reason:     reason,
	}
	if err := billingRepo.LogCreditTransaction(tx); err != nil {
		log.Printf("Warning: refund deduction succeeded but failed to log transaction: %v", err)
		return err
	}

	return nil
}

func (b *Billing) CreateSubscription(userRepo user.UserRepo, subCreatedAttrs SubscriptionAttributes, subscriptionID string) error {
	user, err := userRepo.GetUserByEmail(subCreatedAttrs.UserEmail)
	if err != nil {
		log.Printf("repo.GetUserByEmail failed: %v", err)
		return err
	}

	var tier string
	switch subCreatedAttrs.VariantID {
	case b.VariantIDPro:
		tier = "pro"
	case b.VariantIDPremium:
		tier = "premium"
	default:
		log.Printf("ERROR: unknown variantID: %d", subCreatedAttrs.VariantID)
		return fmt.Errorf("unknown variant ID: %d", subCreatedAttrs.VariantID)
	}

	err = userRepo.UpdateSubscriptionData(
		user.ID,
		"active",
		tier,
		subscriptionID,
		subCreatedAttrs.StartsAt,
		subCreatedAttrs.EndsAt,
	)
	if err != nil {
		log.Printf("CreateSubscriptionData failed: %v", err)
		return err
	}

	return nil
}

func (b *Billing) CancelSubscription(userRepo user.UserRepo, email string) error {
	user, err := userRepo.GetUserByEmail(email)
	if err != nil {
		log.Printf("repo.GetUserByEmail failed: %v", err)
		return err
	}

	err = userRepo.UpdateSubscriptionStatusData(
		user.ID,
		"cancelled",
	)
	if err != nil {
		log.Printf("CancelSubscriptionData failed: %v", err)
		return err
	}

	return nil
}

func (b *Billing) ResumeSubscription(userRepo user.UserRepo, email string) error {
	user, err := userRepo.GetUserByEmail(email)
	if err != nil {
		log.Printf("repo.GetUserByEmail failed: %v", err)
		return err
	}

	err = userRepo.UpdateSubscriptionStatusData(
		user.ID,
		"active",
	)
	if err != nil {
		log.Printf("CancelSubscriptionData failed: %v", err)
		return err
	}

	return nil
}

func (b *Billing) ExpireSubscription(userRepo user.UserRepo, billingRepo BillingRepo, email string) error {
	user, err := userRepo.GetUserByEmail(email)
	if err != nil {
		log.Printf("repo.GetUserByEmail failed: %v", err)
		return err
	}

	err = userRepo.UpdateSubscriptionStatusData(
		user.ID,
		"expired",
	)
	if err != nil {
		log.Printf("CancelSubscriptionData failed: %v", err)
		return err
	}

	if user.SubscriptionCredits > 0 {
		err = userRepo.AddCredits(user.ID, -user.SubscriptionCredits, "subscription")
		if err != nil {
			log.Printf("repo.AddCredits failed: %v", err)
			return err
		}

		tx := CreditTransaction{
			UserID:     user.ID,
			Amount:     -user.SubscriptionCredits,
			CreditType: "subscription",
			Reason:     "Zeroed out credits on subscription expiration",
		}
		if err := billingRepo.LogCreditTransaction(tx); err != nil {
			log.Printf("Warning: zero-out succeeded but failed to log transaction: %v", err)
		}
	}

	return nil
}

func (b *Billing) RenewSubscription(userRepo user.UserRepo, billingRepo BillingRepo, subRenewAttrs SubscriptionRenewAttributes) error {
	user, err := userRepo.GetUserByEmail(subRenewAttrs.UserEmail)
	if err != nil {
		log.Printf("repo.GetUserByEmail failed: %v", err)
		return err
	}

	if subRenewAttrs.Total != 1999 && subRenewAttrs.Total != 2999 {
		return nil
	}

	var (
		credits int
		reason  string
	)
	switch user.SubscriptionTier {
	case "pro":
		credits = 10
		reason = "Pro subscription monthly credit"
	case "premium":
		credits = 20
		reason = "Premium subscription monthly credit"
	default:
		log.Printf("ERROR: unknown user.SubscriptionTier: %s", user.SubscriptionTier)
		return fmt.Errorf("unknown user.SubscriptionTier: %s", user.SubscriptionTier)
	}

	if err := userRepo.AddCredits(user.ID, credits, "subscription"); err != nil {
		log.Printf("repo.AddCredits failed: %v", err)
		return err
	}

	tx := CreditTransaction{
		UserID:     user.ID,
		Amount:     credits,
		CreditType: "subscription",
		Reason:     reason,
	}
	if err := billingRepo.LogCreditTransaction(tx); err != nil {
		log.Printf("Warning: credit granted but failed to log transaction: %v", err)
		return err
	}

	return nil
}

func (b *Billing) ChangeSubscription(userRepo user.UserRepo, billingRepo BillingRepo, subChangedAttrs SubscriptionAttributes) error {
	user, err := userRepo.GetUserByEmail(subChangedAttrs.UserEmail)
	if err != nil {
		log.Printf("repo.GetUserByEmail failed: %v", err)
		return err
	}

	var (
		credits int
		reason  string
	)

	switch subChangedAttrs.VariantID {
	case b.VariantIDPro:
		currentCredits := user.SubscriptionCredits
		if currentCredits >= 10 {
			credits = -10
		} else {
			credits = -currentCredits
		}

		reason = "Premium downgraded to Pro subscription monthly credit"
	case b.VariantIDPremium:
		credits = 10
		reason = "Pro upgraded to Premium subscription monthly credit"
	default:
		log.Printf("ERROR: unknown user.SubscriptionTier: %s", user.SubscriptionTier)
		return fmt.Errorf("unknown user.SubscriptionTier: %s", user.SubscriptionTier)
	}

	if err := userRepo.AddCredits(user.ID, credits, "subscription"); err != nil {
		log.Printf("repo.AddCredits failed: %v", err)
		return err
	}

	tx := CreditTransaction{
		UserID:     user.ID,
		Amount:     credits,
		CreditType: "subscription",
		Reason:     reason,
	}
	if err := billingRepo.LogCreditTransaction(tx); err != nil {
		log.Printf("Warning: credit granted but failed to log transaction: %v", err)
		return err
	}

	return nil
}

func (b *Billing) UpdateSubscription(userRepo user.UserRepo, subUpdatedAttrs SubscriptionAttributes, subscriptionID string) error {
	user, err := userRepo.GetUserByEmail(subUpdatedAttrs.UserEmail)
	if err != nil {
		log.Printf("repo.GetUserByEmail failed: %v", err)
		return err
	}

	var tier string
	switch subUpdatedAttrs.VariantID {
	case b.VariantIDPro:
		tier = "pro"
	case b.VariantIDPremium:
		tier = "premium"
	default:
		log.Printf("ERROR: unknown variantID: %d", subUpdatedAttrs.VariantID)
		return fmt.Errorf("unknown variant ID: %d", subUpdatedAttrs.VariantID)
	}

	err = userRepo.UpdateSubscriptionData(
		user.ID,
		subUpdatedAttrs.Status,
		tier,
		subscriptionID,
		subUpdatedAttrs.StartsAt,
		subUpdatedAttrs.EndsAt,
	)
	if err != nil {
		log.Printf("UpdateSubscriptionData failed: %v", err)
		return err
	}

	return nil
}
