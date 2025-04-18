package user

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateUser(t *testing.T) {

	tests := []struct {
		name        string
		username    string
		email       string
		password    string
		userResp    *User
		expectError bool
	}{
		{
			name:     "UserCreate_Success",
			username: "test",
			email:    "test@test.com",
			password: "test",
			userResp: &User{
				ID:        1,
				Username:  "test",
				Email:     "test@test.com",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewMockRepo()

			user, err := CreateUser(repo, tc.username, tc.email, tc.password)

			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}

			if !tc.expectError {
				expected := tc.userResp
				got := user

				if diff := cmp.Diff(expected, got,
					cmpopts.IgnoreFields(User{}, "Password"),
					cmpopts.EquateApproxTime(time.Second),
				); diff != "" {
					t.Errorf("User mismatch (-want +got):\n%s", diff)
				}

				if err := bcrypt.CompareHashAndPassword(got.Password, []byte(tc.password)); err != nil {
					t.Error("Password hash does not match original password")
				}
			}
		})
	}
}
