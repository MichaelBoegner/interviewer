package testutil

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/token"
)

func CreateTestUserAndJWT() (string, int) {
	var (
		jwt    string
		userID int
	)
	//test user create
	reqBodyUser := strings.NewReader(`{
		"username":"test",
		"email":"test@test.com",
		"password":"test"
	}`)

	_, err := testRequests("", "", "POST", TestServerURL+"/api/users", reqBodyUser)
	if err != nil {
		log.Printf("CreateTestUserAndJWT user creation failed: %v", err)
	}

	//test jwt retrieve
	reqBodyLogin := strings.NewReader(`
		{
			"username": "test",
			"password": "test"
		}
	`)

	resp, err := testRequests("", "", "POST", TestServerURL+"/api/auth/login", reqBodyLogin)
	if err != nil {
		log.Printf("CreateTestUserAndJWT JWT creation failed: %v", err)
	}

	returnVals := &handlers.ReturnVals{}
	json.Unmarshal(resp, returnVals)

	jwt = returnVals.JWToken

	//test userID extract
	userID, err = token.ExtractUserIDFromToken(jwt)
	if err != nil {
		log.Printf("CreateTestUserandJWT userID extraction failed: %v", err)
	}

	return jwt, userID
}

func CreateTestInterview(jwt string) int {
	reqBodyInterview := strings.NewReader(`{}`)

	resp, err := testRequests("Authorization", "Bearer "+jwt, "POST", TestServerURL+"/api/interviews", reqBodyInterview)
	if err != nil {
		log.Printf("CreateTestUserAndJWT JWT creation failed: %v", err)
		return 0
	}

	returnVals := &handlers.ReturnVals{}
	json.Unmarshal(resp, returnVals)

	return returnVals.InterviewID
}

func CreateTestConversation(jwt string, interviewID int) int {
	reqBodyConversation := strings.NewReader(`{
				"conversation_id" : 1,
				"message" : "Answer1"
			}`)
	reqURL := TestServerURL + fmt.Sprintf("/api/conversations/create/%d", interviewID)

	resp, err := testRequests("Authorization", "Bearer "+jwt, "POST", reqURL, reqBodyConversation)
	if err != nil {
		log.Printf("CreateTestUserAndJWT JWT creation failed: %v", err)
		return 0
	}

	returnVals := &handlers.ReturnVals{}
	json.Unmarshal(resp, returnVals)

	return returnVals.Conversation.InterviewID
}

func CreateTestExpiredJWT(id, expires int) string {
	var token *jwt.Token
	jwtSecret := os.Getenv("JWT_SECRET")
	key := []byte(jwtSecret)
	now := time.Now().UTC()

	if expires == 0 {
		expires = 36000
	}
	expiresAt := now.Add(time.Duration(expires) * time.Second)

	claims := jwt.RegisteredClaims{
		Issuer:    "interviewer",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Subject:   strconv.Itoa(id),
	}
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, err := token.SignedString(key)
	if err != nil {
		log.Printf("SignedString failed: %s", err)
		return ""
	}

	return s
}

func TruncateAllTables(db *sql.DB) error {
	_, err := db.Exec(`
		TRUNCATE users, interviews, conversations RESTART IDENTITY CASCADE;
	`)
	return err
}

func testRequests(headerKey, headerValue, method, url string, reqBody *strings.Reader) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		log.Printf("CreateTestUserAndJWT user creation failed: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if headerKey != "" {
		req.Header.Set(headerKey, headerValue)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Reading response failed: %v", err)
		return nil, err
	}

	return bodyBytes, nil
}
