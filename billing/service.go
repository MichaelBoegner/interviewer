package billing

import (
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
)

func (s *Billing) CreateCheckoutSession(userEmail string, priceID string) (string, error) {
	cus, err := customer.New(&stripe.CustomerParams{
		Email: stripe.String(userEmail),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create stripe customer: %w", err)
	}

	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(cus.ID),
		SuccessURL: stripe.String(os.Getenv("FRONTEND_URL") + "payment-success"),
		CancelURL:  stripe.String(os.Getenv("FRONTEND_URL") + "payment-cancel"),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems:  []*stripe.CheckoutSessionLineItemParams{{Price: stripe.String(priceID), Quantity: stripe.Int64(1)}},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return sess.URL, nil
}
