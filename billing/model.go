package billing

import (
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
	Meta MetaData `json:"meta"`
	Data Data     `json:"data"`
}

type MetaData struct {
	EventName string `json:"event_name"`
}

type Data struct {
	Attributes Attributes `json:"attributes"`
}

type Attributes struct {
	CustomerID     string     `json:"customer_id"`
	SubscriptionID string     `json:"id"` // inside attributes in some events
	Status         string     `json:"status"`
	RenewsAt       *time.Time `json:"renews_at"`
	EndsAt         *time.Time `json:"ends_at"`
	TrialEndsAt    *time.Time `json:"trial_ends_at"`
	VariantID      int        `json:"variant_id"`
	UnitPrice      int        `json:"unit_price"` // sometimes inside items array
	Currency       string     `json:"currency"`
}

func NewBilling() *Billing {
	return &Billing{
		APIKey: os.Getenv("BILLING_SECRET_KEY"),
	}
}
