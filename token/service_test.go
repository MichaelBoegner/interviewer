package token

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCreateRefreshToken(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		failRepo    bool
		expectError bool
	}{
		{
			name:        "CreateRefreshToken_Success",
			userID:      1,
			expectError: false,
		},
		{
			name:        "CreateRefreshToken_RepoError",
			userID:      1,
			failRepo:    true,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			repo := NewMockRepo()
			if tc.failRepo {
				repo.failRepo = true
			}

			token, err := CreateRefreshToken(repo, tc.userID)

			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}

			if !tc.expectError && token == "" {
				t.Errorf("expected non-empty token but got empty string")
			}
		})
	}
}

func TestGetStoredRefreshToken(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		storedToken string
		failRepo    bool
		expectError bool
	}{
		{
			name:        "GetStoredRefreshToken_Success",
			userID:      1,
			storedToken: "abc123",
			expectError: false,
		},
		{
			name:        "GetStoredRefreshToken_RepoError",
			userID:      1,
			failRepo:    true,
			expectError: true,
		},
		{
			name:        "GetStoredRefreshToken_NotFound",
			userID:      99,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			repo := NewMockRepo()
			if tc.failRepo {
				repo.failRepo = true
			}

			token, err := GetStoredRefreshToken(repo, tc.userID)

			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}

			if !tc.expectError {
				expected := tc.storedToken
				got := token

				if diff := cmp.Diff(expected, got); diff != "" {
					t.Errorf("Token mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestVerifyRefreshToken(t *testing.T) {
	tests := []struct {
		name        string
		storedToken string
		inputToken  string
		expected    bool
	}{
		{
			name:        "VerifyRefreshToken_Match",
			storedToken: "abc123",
			inputToken:  "abc123",
			expected:    true,
		},
		{
			name:        "VerifyRefreshToken_NoMatch",
			storedToken: "abc123",
			inputToken:  "xyz789",
			expected:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			result := VerifyRefreshToken(tc.storedToken, tc.inputToken)

			if result != tc.expected {
				t.Errorf("expected %v but got %v", tc.expected, result)
			}
		})
	}
}

func TestExtractUserIDFromToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")

	tests := []struct {
		name        string
		userID      int
		invalid     bool
		expectError bool
	}{
		{
			name:   "ExtractUserIDFromToken_Success",
			userID: 42,
		},
		{
			name:        "ExtractUserIDFromToken_InvalidToken",
			invalid:     true,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			var token string
			var err error

			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			if tc.invalid {
				token = "invalid.token.value"
			} else {
				token, err = CreateJWT(tc.userID, 3600)
				if err != nil {
					t.Fatalf("failed to create JWT: %v", err)
				}
			}

			uid, err := ExtractUserIDFromToken(token)

			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}

			if !tc.expectError {
				expected := tc.userID
				got := uid

				if diff := cmp.Diff(expected, got); diff != "" {
					t.Errorf("UserID mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func showLogsIfFail(t *testing.T, name string, buf strings.Builder) {
	log.SetOutput(os.Stderr)
	if t.Failed() {
		fmt.Printf("---- logs for test: %s ----\n%s\n", name, buf.String())
	}
}
