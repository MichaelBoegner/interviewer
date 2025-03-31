package testutil

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/michaelboegner/interviewer/token"
)

func CreateTestUserAndJWT() (string, int) {
	var (
		jwt    string
		userID int
	)
	//test user created
	reqBodyUser := strings.NewReader(`{
		"username":"test",
		"email":"test@email.com",
		"password":"test"
	}`)

	userResp, _, err := testRequests("POST", TestServerURL+"/api/users/", reqBodyUser)
	if err != nil {
		log.Fatalf("CreateTestUserAndJWT user creation failed: %v", err)
	}

	type UserResponse struct {
		UserID   int    `json:"user_id"`
		Username string `json:"username"`
	}
	var user = &UserResponse{}
	json.Unmarshal(userResp, user)

	//test jwt retrieved
	reqBodyLogin := strings.NewReader(`
		{
			"username": "test",
			"password": "test"
		}
	`)

	loginResp, _, err := testRequests("POST", TestServerURL+"/api/auth/login", reqBodyLogin)
	if err != nil {
		log.Fatalf("CreateTestUserAndJWT JWT creation failed: %v", err)
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
		log.Fatalf("CreateTestUserandJWT userID extraction failed: %v", err)
	}

	return jwt, userID
}

func testRequests(method, url string, reqBody *strings.Reader) ([]byte, int, error) {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		log.Fatalf("CreateTestUserAndJWT user creation failed: %v", err)
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request to create test user failed: %v", err)
		return nil, resp.StatusCode, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Reading response failed: %v", err)
		return nil, resp.StatusCode, err
	}

	return bodyBytes, resp.StatusCode, nil
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
		log.Fatalf("Bad SignedString: %s", err)
		return ""
	}

	return s
}
