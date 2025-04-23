package billing

import (
	"os"

	"github.com/stripe/stripe-go/v76"
)

type Billing struct{}

func NewBilling() Billing {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	return Billing{}
}
