package billing

import (
	"os"
)

type Billing struct {
	APIKey string
}

type CheckoutPayload struct {
	VariantID int    `json:"variant_id"`
	Email     string `json:"checkout_data[email]"`
}

type CheckoutResponse struct {
	Data struct {
		Attributes struct {
			URL string `json:"url"`
		} `json:"attributes"`
	} `json:"data"`
}

func NewBilling() *Billing {
	return &Billing{
		APIKey: os.Getenv("STRIPE_SECRET_KEY"),
	}
}
