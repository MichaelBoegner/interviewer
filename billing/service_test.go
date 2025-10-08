package billing_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/michaelboegner/interviewer/billing"
	"github.com/michaelboegner/interviewer/user"
)

func NewTestBilling() *billing.Billing {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	return &billing.Billing{
		VariantIDIndividual: 1,
		VariantIDPro:        2,
		VariantIDPremium:    3,
		Logger:              logger,
	}
}

func TestApplyCredits(t *testing.T) {
	tests := []struct {
		name       string
		variantID  int
		expectErr  bool
		failUser   bool
		failCredit bool
		failLog    bool
	}{
		{
			name:       "ApplyCredits_Individual_Success",
			variantID:  1,
			expectErr:  false,
			failUser:   false,
			failCredit: false,
			failLog:    false,
		},
		{
			name:       "ApplyCredits_Pro_Success",
			variantID:  2,
			expectErr:  false,
			failUser:   false,
			failCredit: false,
			failLog:    false,
		},
		{
			name:       "ApplyCredits_UnknownVariant",
			variantID:  999,
			expectErr:  true,
			failUser:   false,
			failCredit: false,
			failLog:    false,
		},
		{
			name:       "ApplyCredits_UserRepoFail",
			variantID:  1,
			expectErr:  true,
			failUser:   true,
			failCredit: false,
			failLog:    false,
		},
		{
			name:       "ApplyCredits_AddCreditsFail",
			variantID:  1,
			expectErr:  true,
			failUser:   false,
			failCredit: true,
			failLog:    false,
		},
		{
			name:       "ApplyCredits_LogFail",
			variantID:  1,
			expectErr:  true,
			failUser:   false,
			failCredit: false,
			failLog:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			userRepo := user.NewMockRepo()
			billingRepo := billing.NewMockRepo()

			userRepo.FailGetUserByEmail = tc.failUser
			userRepo.FailAddCredits = tc.failCredit
			billingRepo.FailLogCreditTransaction = tc.failLog

			b := NewTestBilling()

			err := b.ApplyCredits("test@example.com", tc.variantID)
			if tc.expectErr && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}
		})
	}
}

func TestDeductCredits(t *testing.T) {
	tests := []struct {
		name       string
		variantID  int
		expectErr  bool
		failUser   bool
		failCredit bool
		failLog    bool
	}{
		{
			name:       "DeductCredits_Pro_Success",
			variantID:  2,
			expectErr:  false,
			failUser:   false,
			failCredit: false,
			failLog:    false,
		},
		{
			name:       "DeductCredits_UnknownVariant",
			variantID:  999,
			expectErr:  true,
			failUser:   false,
			failCredit: false,
			failLog:    false,
		},
		{
			name:       "DeductCredits_UserRepoFail",
			variantID:  2,
			expectErr:  true,
			failUser:   true,
			failCredit: false,
			failLog:    false,
		},
		{
			name:       "DeductCredits_AddCreditsFail",
			variantID:  2,
			expectErr:  true,
			failUser:   false,
			failCredit: true,
			failLog:    false,
		},
		{
			name:       "DeductCredits_LogFail",
			variantID:  2,
			expectErr:  true,
			failUser:   false,
			failCredit: false,
			failLog:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			userRepo := user.NewMockRepo()
			billingRepo := billing.NewMockRepo()

			userRepo.FailGetUserByEmail = tc.failUser
			userRepo.FailAddCredits = tc.failCredit
			billingRepo.FailLogCreditTransaction = tc.failLog

			b := NewTestBilling()

			attrs := billing.OrderAttributes{
				UserEmail: "test@example.com",
				FirstOrderItem: struct {
					VariantID int `json:"variant_id"`
				}{
					VariantID: tc.variantID,
				},
			}

			err := b.DeductCredits(attrs)
			if tc.expectErr && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}
		})
	}
}

func TestVerifyBillingSignature(t *testing.T) {
	b := NewTestBilling()
	body := []byte(`{"key":"value"}`)
	secret := "testsecret"
	mac := hmacSha256(body, secret)
	if !b.VerifyBillingSignature(mac, body, secret) {
		t.Fatal("expected signature to be valid")
	}
	if b.VerifyBillingSignature("invalid", body, secret) {
		t.Fatal("expected signature to be invalid")
	}
}

func hmacSha256(message []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(message)
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func showLogsIfFail(t *testing.T, name string, buf strings.Builder) {
	if t.Failed() {
		t.Logf("---- logs for test: %s ----\n%s\n", name, buf.String())
	}
}
