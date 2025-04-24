package mailer

import "os"

type Mailer struct {
	APIKey  string
	BaseURL string
}

func NewMailer() *Mailer {
	return &Mailer{
		APIKey:  os.Getenv("RESEND_API_KEY"),
		BaseURL: "https://api.resend.com",
	}
}
