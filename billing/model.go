package billing

import (
	"encoding/json"
	"os"
	"time"
)

type Billing struct {
	APIKey string
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

func NewBilling() *Billing {
	return &Billing{
		APIKey: os.Getenv("BILLING_SECRET_KEY"),
	}
}
