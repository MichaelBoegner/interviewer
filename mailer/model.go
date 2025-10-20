package mailer

import (
	"log/slog"
	"os"
)

type MailerService struct {
	APIKey  string
	BaseURL string
	Logger  *slog.Logger
}

const signature = `
		<br><br>
		<p>
			<strong>Michael Boegner</strong><br>
			Founder • <a href="https://interviewer.dev" style="color: #007bff; text-decoration: none;">Interviewer</a><br>
			<a href="mailto:support@mail.interviewer.dev" style="color: #000;">support@mail.interviewer.dev</a><br>
		</p>
		<p style="color: gray; font-size: 12px; margin-top: 4px;">
			Everything gets easier with practice!
		</p>
	`

func NewMailerService(logger *slog.Logger) *MailerService {
	return &MailerService{
		APIKey:  os.Getenv("RESEND_API_KEY"),
		BaseURL: "https://api.resend.com",
		Logger:  logger,
	}
}

type MailerClient interface {
	SendPasswordReset(email, resetURL string) error
	SendVerificationEmail(email, verifyURL string) error
	SendWelcome(email string) error
	SendDeletionConfirmation(email string) error
}
