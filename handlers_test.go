package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/user"
)

func TestUsersHandler(t *testing.T) {
	// Mock repository logic
	mockUserRepo := user.NewMockRepo()

	apiCfg := &apiConfig{UserRepo: mockUserRepo}

	// Simulate a request
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"testuser", "email": "test1@example.com", "password":"password"}`))
	req = req.WithContext(context.WithValue(req.Context(), "params", middleware.AcceptedVals{
		Username: "testuser", Email: "test1@example.com", Password: "password",
	}))
	w := httptest.NewRecorder()

	// Call the handler
	apiCfg.usersHandler(w, req)

	// Assert the response
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
