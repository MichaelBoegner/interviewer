package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/michaelboegner/interviewer/token"
)

func CreateTestUserAndJWT(t *testing.T) (string, string, int) {
	t.Helper()

	var (
		username string
		jwt      string
		userID   int
	)
	//test user created
	reqBodyUser := strings.NewReader(`{
		"username":"test",
		"email":"test@email.com",
		"password":"test"
	}`)

	userResp, err := testRequests(t, "POST", TestServerURL+"/api/users/", reqBodyUser)
	if err != nil {
		t.Fatalf("CreateTestUserAndJWT user creation failed: %v", err)
	}

	type UserResponse struct {
		UserID   int    `json:"user_id"`
		Username string `json:"username"`
	}
	var user = &UserResponse{}
	json.Unmarshal(userResp, user)
	username = user.Username

	//test jwt retrieved
	reqBodyLogin := strings.NewReader(`
		{
			"username": "test",
			"password": "test"
		}
	`)

	loginResp, err := testRequests(t, "POST", TestServerURL+"/api/auth/login", reqBodyLogin)
	if err != nil {
		t.Fatalf("CreateTestUserAndJWT JWT creation failed: %v", err)
	}

	type AuthResponse struct {
		UserID       int    `json:"user_id"`
		Username     string `json:"username"`
		JWToken      string `json:"jwtoken"`
		RefreshToken string `json:"refresh_token"`
	}

	var decodedLoginResp = &AuthResponse{}
	json.Unmarshal(loginResp, decodedLoginResp)

	jwt = decodedLoginResp.JWToken

	//test userID extracted
	userID, err = token.ExtractUserIDFromToken(jwt)
	if err != nil {
		t.Fatalf("CreateTestUserandJWT userID extraction failed: %v", err)
	}

	return username, jwt, userID
}

func testRequests(t *testing.T, method, url string, reqBody *strings.Reader) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Logf("CreateTestUserAndJWT user creation failed: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Request to create test user failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Reading response failed: %v", err)
		return nil, err
	}

	return bodyBytes, nil
}
