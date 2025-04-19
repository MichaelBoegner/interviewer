package user

import (
	"testing"

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
		user        *User
		failRepo    bool
		expectError bool
	}{
		{
			name:     "CreateUser_Success",
			username: "test",
			email:    "test@test.com",
			password: "test",
			user: &User{
				ID:       1,
				Username: "test",
				Email:    "test@test.com",
			},
			expectError: false,
		},
		{
			name:        "CreateUser_RepoError",
			username:    "test",
			email:       "test@test.com",
			password:    "test",
			user:        nil,
			failRepo:    true,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewMockRepo()
			if tc.failRepo {
				repo.failRepo = true
			}

			user, err := CreateUser(repo, tc.username, tc.email, tc.password)

			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}

			if !tc.expectError {
				expected := tc.user
				got := user

				if diff := cmp.Diff(expected, got,
					cmpopts.IgnoreFields(User{}, "Password", "CreatedAt", "UpdatedAt"),
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

func TestLoginUser(t *testing.T) {

	tests := []struct {
		name        string
		username    string
		password    string
		jwtoken     string
		userID      int
		failRepo    bool
		expectError bool
	}{
		{
			name:        "LoginUser_Success",
			username:    "test",
			password:    "test",
			userID:      1,
			expectError: false,
		},
		{
			name:        "LoginUser_RepoError",
			username:    "test",
			password:    "test",
			failRepo:    true,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewMockRepo()
			if tc.failRepo {
				repo.failRepo = true
			}

			jwtoken, userID, err := LoginUser(repo, tc.username, tc.password)

			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}

			if !tc.expectError {
				expected := tc.userID
				got := userID

				if diff := cmp.Diff(expected, got); diff != "" {
					t.Errorf("User mismatch (-want +got):\n%s", diff)
				}
				if jwtoken == "" {
					t.Errorf("Expected jwtoken but got empty string")
				}
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		user        *User
		failRepo    bool
		expectError bool
	}{
		{
			name:   "GetUser_Success",
			userID: 1,
			user: &User{
				ID:       1,
				Username: "test",
				Email:    "test@test.com",
			},
			expectError: false,
		},
		{
			name:        "GetUser_RepoError",
			userID:      1,
			failRepo:    true,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewMockRepo()
			if tc.failRepo {
				repo.failRepo = true
			}

			user, err := GetUser(repo, tc.userID)

			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}

			if !tc.expectError {
				expected := tc.user
				got := user

				if diff := cmp.Diff(expected, got,
					cmpopts.IgnoreFields(User{}, "Password", "CreatedAt", "UpdatedAt"),
				); diff != "" {
					t.Errorf("User mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
