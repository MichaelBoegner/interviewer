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

func (b *Billing) CreateCheckoutSession(userEmail string, variantID int) (string, error) {
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

func (b *Billing) DeleteSubscription(subscriptionID string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	//DEBUG
	fmt.Printf("subscriptionid: %v", subscriptionID)
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

func (b *Billing) ExpireSubscription(userRepo user.UserRepo, email string) error {
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
