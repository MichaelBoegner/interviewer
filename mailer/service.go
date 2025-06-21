package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (m *Mailer) SendPasswordReset(email, resetURL string) error {
	payload := map[string]any{
		"from":    "Interviewer Support <support@mail.interviewer.dev>",
		"to":      email,
		"subject": "Reset your password",
		"html": fmt.Sprintf(`
	<p>
		We received a request to reset your password.<br><br>
		If you made this request, click <a href="%s">here</a> to reset your password.<br><br>
		If you didn’t request a password reset, you can safely ignore this email.
	</p>`, resetURL) + signature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Marshal failed: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/emails", m.BaseURL), bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Mailer NewRequest failed: %v", err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Mailer Client Do failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Resend error: %s", resp.Status)
	}

	return nil
}

func (m *Mailer) SendVerificationEmail(email, verifyURL string) error {

	payload := map[string]any{
		"from":    "Interviewer Support <support@mail.interviewer.dev>",
		"to":      email,
		"subject": "Verify your Interviewer account",
		"html": fmt.Sprintf(`
	<p>
		Hey there!<br><br>
		Thanks for signing up for Interviewer. We're excited to help you prep for your next big opportunity!<br><br>
		Click <a href="%s">here</a> to verify your account and get started.
	</p>`, verifyURL) + signature,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", m.BaseURL+"/emails", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+m.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend error: %s", resp.Status)
	}
	return nil
}

func (m *Mailer) SendWelcome(email string) error {
	payload := map[string]any{
		"from":    "Interviewer Support <support@mail.interviewer.dev>",
		"to":      email,
		"subject": "Welcome to Interviewer!",
		"html": `
<p>
	Your Interviewer account has been successfully verified!
</p>
<p>
	To help you get started, we've added <strong>one free interview</strong> to your account.
</p>
<p>
	Head over to your dashboard to begin your first mock interview.
</p>
<div style="margin-top: 30px;">
	<a href="https://interviewer.dev/dashboard" style="
		background-color: #4CAF50;
		color: white;
		padding: 12px 24px;
		text-decoration: none;
		border-radius: 4px;
		display: inline-block;
		font-size: 16px;
		font-family: sans-serif;
	">
		Go to Dashboard
	</a>
</div>
` + signature,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", m.BaseURL+"/emails", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+m.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend error: %s", resp.Status)
	}
	return nil
}

func (m *Mailer) SendDeletionConfirmation(email string) error {
	payload := map[string]any{
		"from":    "Interviewer Support <support@mail.interviewer.dev>",
		"to":      email,
		"subject": "Your Interviewer account has been deleted",
		"html": `
<p>
	We're sorry to see you go, but your Interviewer account has been successfully deleted.
</p>
<p>
	Any remaining interview credits have been deactivated, and if you had an active subscription, it has been fully canceled. You will not be charged again.
</p>
<p>
	If you change your mind, you're always welcome to create a new account at any time.
</p>
<p>
	Thanks again for giving Interviewer a try — we genuinely appreciate it and wish you all the best in your interview journey.
</p>
` + signature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Marshal failed: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/emails", m.BaseURL), bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Mailer NewRequest failed: %v", err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Mailer Client Do failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Resend error: %s", resp.Status)
	}

	return nil
}
