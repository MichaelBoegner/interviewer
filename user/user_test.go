package user_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/michaelboegner/interviewer/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserIntegration(t *testing.T) {
	reqBody, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"email":    "testuser@example.com",
		"password": "securepassword123",
	})

	resp, err := http.Post(testutil.TestServerURL+"/api/users/", "application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var responseData map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseData)

	// assert.NotNil(t, responseData["id"])
	assert.Equal(t, "testuser", responseData["username"])
}

func TestMain(m *testing.M) {
	log.Println("Loading environment variables...")
	err := godotenv.Load("../.env.test")
	if err != nil {
		log.Fatalf("Error loading .env.test file: %v", err)
	}

	log.Println("Initializing test server...")
	testutil.InitTestServer()

	// 🚨 Check `TestServerURL` before running any tests
	if testutil.TestServerURL == "" {
		log.Fatal("TestMain: TestServerURL is empty! The server did not start properly.")
	}

	log.Printf("TestMain: Test server started successfully at: %s", testutil.TestServerURL)

	code := m.Run()

	log.Println("Stopping test server...")
	testutil.StopTestServer()

	os.Exit(code)
}
