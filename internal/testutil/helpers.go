package testutil

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

func CreateTestUserAndJWT(logger *slog.Logger) (string, int) {
	var (
		jwt    string
		userID int
	)
	//test user create
	username := "test"
	email := "test@test.com"
	password := "test"

	verificationJWT, err := user.VerificationToken(email, username, password)
	if err != nil {
		logger.Error("GenerateEmailVerificationToken failed", "error", err)
	}

	reqBodyUser := strings.NewReader(fmt.Sprintf(`{
			"token": "%s"
		}`, verificationJWT))

	_, err = testRequests("", "", "POST", TestServerURL+"/api/users", reqBodyUser, logger)
	if err != nil {
		logger.Error("CreateTestUserAndJWT user creation failed", "error", err)
	}

	//test jwt retrieve
	reqBodyLogin := strings.NewReader(`
		{
			"email": "test@test.com",
			"password": "test"
		}
	`)

	resp, err := testRequests("", "", "POST", TestServerURL+"/api/auth/login", reqBodyLogin, logger)
	if err != nil {
		logger.Error("CreateTestUserAndJWT JWT creation failed", "error", err)
	}

	returnVals := &handlers.ReturnVals{}
	json.Unmarshal(resp, returnVals)

	jwt = returnVals.JWToken

	//test userID extract
	userID, err = token.ExtractUserIDFromToken(jwt)
	if err != nil {
		logger.Error("CreateTestUserandJWT userID extraction failed", "error", err)
	}

	return jwt, userID
}

func CreateTestInterview(jwt string, logger *slog.Logger) int {
	reqBodyInterview := strings.NewReader(`{}`)

	resp, err := testRequests("Authorization", "Bearer "+jwt, "POST", TestServerURL+"/api/interviews", reqBodyInterview, logger)
	if err != nil {
		logger.Error("CreateTestUserAndJWT JWT creation failed", "error", err)
		return 0
	}

	returnVals := &handlers.ReturnVals{}
	json.Unmarshal(resp, returnVals)

	return returnVals.InterviewID
}

func CreateTestConversation(jwt string, interviewID int, logger *slog.Logger) int {
	reqBodyConversation := strings.NewReader(`{
				"conversation_id" : 1,
				"message" : "T1Q1A1"
			}`)
	reqURL := TestServerURL + fmt.Sprintf("/api/conversations/create/%d", interviewID)

	resp, err := testRequests("Authorization", "Bearer "+jwt, "POST", reqURL, reqBodyConversation, logger)
	if err != nil {
		logger.Error("CreateTestUserAndJWT JWT creation failed", "error", err)
		return 0
	}

	returnVals := &handlers.ReturnVals{}
	json.Unmarshal(resp, returnVals)

	return returnVals.Conversation.ID
}

func CreateTestExpiredJWT(id, expires int, logger *slog.Logger) string {
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
		logger.Error("SignedString failed", "error", err)
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

func testRequests(headerKey, headerValue, method, url string, reqBody *strings.Reader, logger *slog.Logger) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		logger.Error("CreateTestUserAndJWT user creation failed", "error", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if headerKey != "" {
		req.Header.Set(headerKey, headerValue)
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Request failed", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Reading response failed", "error", err)
		return nil, err
	}

	return bodyBytes, nil
}
