package billing

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Billing struct {
	APIKey              string
	VariantIDIndividual int
	VariantIDPro        int
	VariantIDPremium    int
}

type CheckoutPayload struct {
	VariantID int
	Email     string
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
	} `json:"meta"`

	Data struct {
		Attributes json.RawMessage `json:"attributes"`
	} `json:"data"`
}

type OrderCreatedAttributes struct {
	UserEmail      string `json:"user_email"`
	UserName       string `json:"user_name"`
	TestMode       bool   `json:"test_mode"`
	FirstOrderItem struct {
		ProductID   int    `json:"product_id"`
		VariantID   int    `json:"variant_id"`
		ProductName string `json:"product_name"`
		VariantName string `json:"variant_name"`
	} `json:"first_order_item"`
}

type SubscriptionCancelledAttributes struct {
	UserEmail string     `json:"user_email"`
	UserName  string     `json:"user_name"`
	Status    string     `json:"status"`
	EndsAt    *time.Time `json:"ends_at"`
}

type CreditTransaction struct {
	UserID     int
	Amount     int
	CreditType string
	Reason     string
}

type BillingRepo interface {
	LogCreditTransaction(tx CreditTransaction) error
}

func NewBilling() (*Billing, error) {
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
		APIKey:              os.Getenv("BILLING_SECRET_KEY"),
		VariantIDIndividual: individualID,
		VariantIDPro:        proID,
		VariantIDPremium:    premiumID,
	}, nil
}
