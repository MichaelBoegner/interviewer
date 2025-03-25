package testutil

import (
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
)

func CreateTestUserAndJWT(t *testing.T) (string, string, error) {

	t.Helper()
	client := &http.Client{}

	reqBodyUser := `{
		"username":"test",
		"email":"test@email.com",
		"password":"test"
	}`

	t.Logf("Request body being sent: %s", reqBodyUser)
	req, err := http.NewRequest("POST", TestServerURL+"/api/users/", strings.NewReader(reqBodyUser))
	if err != nil {
		t.Logf("CreateTestUserAndJWT user creation failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Request to create test user failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Reading response failed: %v", err)
	}

	log.Printf("\n\n\nTHIS IS THE RESP: %v\n\n\n", string(bodyBytes))

	// 2. Login via POST /api/auth/login
	// reqBodyLogin := `
	// 	{
	// 		"username": "test",
	// 		"password": "test"
	// 	}
	// `

	// 3. Extract JWT + userID from response

	return "", "", nil
}
