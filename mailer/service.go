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
		"html":    "<p>" + fmt.Sprintf("To reset your password, click here: %s", resetURL) + "</p>",
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
