package testutil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func CreateTestUserAndJWT(t *testing.T, router http.Handler) (int, string) {
	t.Helper()

	// 1. Register a new user via POST /api/users
	reqBody := `{
		"username":"test",
		"email":"test@email.com",
		"password":"test
	}`
	req := httptest.NewRequest("POST", "/api/users", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	// 2. Login via POST /api/auth/login
	// 3. Extract JWT + userID from response
	print(
	return 0, ""
}
