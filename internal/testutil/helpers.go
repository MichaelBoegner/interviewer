package testutil

import (
	"database/sql"
	"encoding/json"
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
		"email":"test@email.com",
		"password":"test"
	}`)

	userResp, err := testRequests("", "", "POST", TestServerURL+"/api/users", reqBodyUser)
	if err != nil {
		log.Printf("CreateTestUserAndJWT user creation failed: %v", err)
	}

	UserUnmarshaled := &handlers.ReturnVals{}
	json.Unmarshal(userResp, UserUnmarshaled)

	//test jwt retrieve
	reqBodyLogin := strings.NewReader(`
		{
			"username": "test",
			"password": "test"
		}
	`)

	loginResp, err := testRequests("", "", "POST", TestServerURL+"/api/auth/login", reqBodyLogin)
	if err != nil {
		log.Printf("CreateTestUserAndJWT JWT creation failed: %v", err)
	}

	loginRespUnmarshaled := &handlers.ReturnVals{}
	json.Unmarshal(loginResp, loginRespUnmarshaled)

	jwt = loginRespUnmarshaled.JWToken

	//test userID extract
	userID, err = token.ExtractUserIDFromToken(jwt)
	if err != nil {
		log.Printf("CreateTestUserandJWT userID extraction failed: %v", err)
	}

	return jwt, userID
}

func CreateTestInterview(jwt string) int {
	reqBodyInterview := strings.NewReader(`{}`)

	interviewResp, err := testRequests("Authorization", "Bearer "+jwt, "POST", TestServerURL+"/api/interviews", reqBodyInterview)
	if err != nil {
		log.Printf("CreateTestUserAndJWT JWT creation failed: %v", err)
		return 0
	}

	interviewRespUnmarshaled := &handlers.ReturnVals{}
	json.Unmarshal(interviewResp, interviewRespUnmarshaled)

	interviewID := interviewRespUnmarshaled.InterviewID

	return interviewID
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

func CreateTestJWT(id, expires int) string {
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
