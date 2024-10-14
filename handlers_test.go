package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/user"
)

func TestUsersHandler_Post(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        string
		params         middleware.AcceptedVals
		expectedStatus int
		expectError    bool
	}{
		{
			name:    "CreateUser_Success",
			reqBody: `{"username":"testuser", "email":"test@example.com", "password":"password"}`,
			params: middleware.AcceptedVals{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "MissingParams",
			reqBody:        `{"username":"", "email":"", "password":""}`,
			params:         middleware.AcceptedVals{},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockUserRepo := user.NewMockRepo()
			apiCfg := &apiConfig{UserRepo: mockUserRepo}

			req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(tc.reqBody))
			req = req.WithContext(context.WithValue(req.Context(), "params", tc.params))
			w := httptest.NewRecorder()

			// Act
			apiCfg.usersHandler(w, req)

			// Assert
			if w.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			var resp returnVals

			err := json.Unmarshal(w.Body.Bytes(), &resp)
			fmt.Printf("resp: %v\n", resp)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			fmt.Printf("resp.Error: %v\n", resp.Error)
			if tc.expectError && resp.Error == "" {
				t.Errorf("expected error message, got none")
			}
		})
	}
}
