package testutil

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func CreateTestUserAndJWT(t *testing.T) (string, string, error) {
	t.Helper()

	var (
		user string
	)

	reqBodyUser := strings.NewReader(`{
		"username":"test",
		"email":"test@email.com",
		"password":"test"
	}`)

	bodyBytes, err := testRequests(t, "POST", TestServerURL+"/api/users/", reqBodyUser)
	if err != nil {
		t.Logf("CreateTestUserAndJWT user creation failed: %v", err)
	}

	user = string(bodyBytes)

	// 2. Login via POST /api/auth/login
	reqBodyLogin := `
		{
			"username": "test",
			"password": "test"
		}
	`

	// 3. Extract JWT + userID from response

	return user, "", nil
}

func testRequests(t *testing.T, method, url string, reqBody *strings.Reader) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, TestServerURL+url, reqBody)
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
