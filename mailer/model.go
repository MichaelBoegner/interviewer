package mailer

import "os"

type Mailer struct {
	APIKey  string
	BaseURL string
}

const signature = `
		<br><br>
		<p>
			<strong>Michael Boegner</strong><br>
			Founder • <a href="https://interviewer.dev" style="color: #007bff; text-decoration: none;">Interviewer</a><br>
			<a href="mailto:support@mail.interviewer.dev" style="color: #000;">support@mail.interviewer.dev</a><br>
		</p>
		<p style="color: gray; font-size: 12px; margin-top: 8px;">
			Everything gets easier with practice!
		</p>
	`

func NewMailer() *Mailer {
	return &Mailer{
		APIKey:  os.Getenv("RESEND_API_KEY"),
		BaseURL: "https://api.resend.com",
	}
}
