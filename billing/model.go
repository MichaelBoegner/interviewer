package billing

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/michaelboegner/interviewer/user"
)

type Billing struct {
	BillingRepo         BillingRepo
	UserRepo            user.UserRepo
	APIKey              string
	VariantIDIndividual int
	VariantIDPro        int
	VariantIDPremium    int
	Logger              *slog.Logger
}
type CheckoutPayload struct {
	Data CheckoutData `json:"data"`
}

type CheckoutData struct {
	Type          string                `json:"type"`
	Attributes    CheckoutAttributes    `json:"attributes"`
	Relationships CheckoutRelationships `json:"relationships"`
}

type CheckoutAttributes struct {
	CheckoutData CheckoutCustomerInfo `json:"checkout_data"`
}

type CheckoutCustomerInfo struct {
	Email string `json:"email"`
}

type CheckoutRelationships struct {
	Store   Relationship `json:"store"`
	Variant Relationship `json:"variant"`
}

type Relationship struct {
	Data RelationshipData `json:"data"`
}

type RelationshipData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type CheckoutResponse struct {
	Data struct {
		Attributes struct {
			URL string `json:"url"`
		} `json:"attributes"`
	} `json:"data"`
}

type BillingWebhookPayload struct {
	Meta struct {
		EventName string `json:"event_name"`
		WebhookID string `json:"webhook_id"`
	} `json:"meta"`

	Data struct {
		SubscriptionID string          `json:"id"`
		Attributes     json.RawMessage `json:"attributes"`
	} `json:"data"`
}

type OrderAttributes struct {
	UserEmail      string `json:"user_email"`
	FirstOrderItem struct {
		VariantID int `json:"variant_id"`
	} `json:"first_order_item"`
}

type SubscriptionAttributes struct {
	UserEmail string    `json:"user_email"`
	Status    string    `json:"status"`
	StartsAt  time.Time `json:"created_at"`
	EndsAt    time.Time `json:"renews_at"`
	VariantID int       `json:"variant_id"`
}

type SubscriptionRenewAttributes struct {
	UserEmail     string `json:"user_email"`
	Total         int    `json:"total"`
	BillingReason string `json:"billing_reason"`
}

type CreditTransaction struct {
	UserID     int
	Amount     int
	CreditType string
	Reason     string
}

type BillingRepo interface {
	LogCreditTransaction(tx CreditTransaction) error
	HasWebhookBeenProcessed(id string) (bool, error)
	MarkWebhookProcessed(id string, event string) error
}

func NewBilling(billingRepo BillingRepo, userRepo user.UserRepo, logger *slog.Logger) (*Billing, error) {
	individualID, err := strconv.Atoi(os.Getenv("LEMON_VARIANT_ID_INDIVIDUAL"))
	if err != nil {
		return nil, fmt.Errorf("invalid INDIVIDUAL ID: %w", err)
	}
	proID, err := strconv.Atoi(os.Getenv("LEMON_VARIANT_ID_PRO"))
	if err != nil {
		return nil, fmt.Errorf("invalid PRO ID: %w", err)
	}
	premiumID, err := strconv.Atoi(os.Getenv("LEMON_VARIANT_ID_PREMIUM"))
	if err != nil {
		return nil, fmt.Errorf("invalid PREMIUM ID: %w", err)
	}
	return &Billing{
		BillingRepo:         billingRepo,
		UserRepo:            userRepo,
		APIKey:              os.Getenv("LEMON_API_KEY"),
		VariantIDIndividual: individualID,
		VariantIDPro:        proID,
		VariantIDPremium:    premiumID,
		Logger:              logger,
	}, nil
}
