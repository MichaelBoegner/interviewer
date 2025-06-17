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
