package customerio

import "os"

type Mailer struct {
	APIKey  string
	APIPass string
	BaseURL string
}

func New(apiKey string) *Mailer {
	return &Mailer{
		APIKey:  os.Getenv("CIO_KEY"),
		APIPass: os.Getenv("CIO_PASS"),
		BaseURL: "https://track.customer.io/api/v1",
	}
}
