package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

type TestCase struct {
	name           string
	reqBody        string
	params         middleware.AcceptedVals
	expectedStatus int
	expectError    bool
	respBody       returnVals
}

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	code := m.Run()

	os.Exit(code)
}

func TestUsersHandler_Post(t *testing.T) {
	tests := []TestCase{
		{
			name:    "CreateUser_Success",
			reqBody: `{"username":"testuser", "email":"test@example.com", "password":"password"}`,
			params: middleware.AcceptedVals{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			respBody: returnVals{
				Username: "testuser",
				Email:    "test@example.com",
			},
		},
		{
			name:           "MissingParams",
			reqBody:        `{"username":"", "email":"", "password":""}`,
			params:         middleware.AcceptedVals{},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			respBody:       returnVals{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockUserRepo := user.NewMockRepo()
			apiCfg := &apiConfig{UserRepo: mockUserRepo}

			w, req, tc, err := setRequestAndWriter(http.MethodPost, "/api/users", tc)
			if err != nil {
				t.Fatalf("failed to set request and writer")
			}

			// Act
			apiCfg.usersHandler(w, req)

			// Assert
			if w.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			// Validate resp
			resp, err := checkResponse(w, tc.respBody, tc.expectError)
			if err != nil {
				t.Fatalf("expected response %v and error %v\ngot response: %v and error %v", tc.respBody, tc.expectError, resp, resp.Error)
			}
		})
	}
}

func TestUsersHandler_Get(t *testing.T) {
	tests := []TestCase{
		{
			name:           "GetUser_Success",
			reqBody:        `{}`,
			params:         middleware.AcceptedVals{},
			expectedStatus: http.StatusOK,
			expectError:    false,
			respBody: returnVals{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockUserRepo := user.NewMockRepo()
			apiCfg := &apiConfig{UserRepo: mockUserRepo}

			w, req, tc, err := setRequestAndWriter(http.MethodGet, "/api/users/1", tc)
			if err != nil {
				t.Fatalf("failed to set request and writer")
			}

			// Act
			apiCfg.usersHandler(w, req)

			// Assert
			if w.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			// Validate resp
			resp, err := checkResponse(w, tc.respBody, tc.expectError)
			if err != nil {
				t.Fatalf("expected response %v and error %v\ngot response: %v and error %v", tc.respBody, tc.expectError, resp, resp.Error)
			}
		})
	}
}

func TestLoginHandler_Post(t *testing.T) {
	tests := []TestCase{
		{
			name:    "LoginUser_Success",
			reqBody: `{"username":"testuser", "password":"password"}`,
			params: middleware.AcceptedVals{
				Username: "testuser",
				Password: "password",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			respBody: returnVals{
				ID:           1,
				JWToken:      "",
				RefreshToken: "",
			},
		},
		{
			name:    "LoginUser_Missing_Username",
			reqBody: `{"username":"", "password":"password"}`,
			params: middleware.AcceptedVals{
				Username: "",
				Password: "password",
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			respBody:       returnVals{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockTokenRepo := token.NewMockRepo()
			mockUserRepo := user.NewMockRepo()
			apiCfg := &apiConfig{
				TokenRepo: mockTokenRepo,
				UserRepo:  mockUserRepo,
			}

			w, req, tc, err := setRequestAndWriter(http.MethodPost, "/api/auth/login", tc)
			if err != nil {
				t.Fatalf("failed to set request and writer")
			}

			// Act
			apiCfg.loginHandler(w, req)

			// Assert
			if w.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			// Validate resp
			resp, err := checkResponse(w, tc.respBody, tc.expectError)
			if err != nil {
				t.Fatalf("expected response %v and error %v\ngot response: %v and error %v", tc.respBody, tc.expectError, resp, resp.Error)
			}

			if !tc.expectError && (resp.JWToken == "" || resp.RefreshToken == "") {
				t.Fatalf("expected non-empty tokens, got empty jwt or refresh token")
			}
		})
	}
}

func TestInterviewsHandler_Post(t *testing.T) {
	token, err := createJWT(1, 0)
	if err != nil || token == "" {
		t.Fatalf("Mock JWT was not created or is empty")
	}

	tests := []TestCase{
		{
			name:    "InterviewsHandler_Success",
			reqBody: `{}`,
			params: middleware.AcceptedVals{
				AccessToken: token,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			respBody:       returnVals{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockUserRepo := user.NewMockRepo()
			mockInterviewRepo := interview.NewMockRepo()

			apiCfg := &apiConfig{
				UserRepo:      mockUserRepo,
				InterviewRepo: mockInterviewRepo,
			}

			w, req, tc, err := setRequestAndWriter(http.MethodPost, "/api/interviews", tc)
			if err != nil {
				t.Fatalf("failed to set request and writer")
			}

			req.Header.Set("Authorization", "Bearer "+token)

			// Apply the middleware to the handler
			handler := middleware.GetContext(http.HandlerFunc(apiCfg.interviewsHandler))

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			if w.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			// Validate resp
			resp, err := checkResponse(w, tc.respBody, tc.expectError)
			if err != nil {
				t.Fatalf("expected response %v and error %v\ngot response: %v and error %v", tc.respBody, tc.expectError, resp, resp.Error)
			}
		})
	}
}

func TestRefreshTokensHandler_Post(t *testing.T) {
	tokenKey := "9942443a086328dfaa867e0708426f94284d25700fa9df930261e341f0d8c671"

	tests := []TestCase{
		{
			name:    "RefreshTokensHandler_Success",
			reqBody: `{"user_id": 1}`,
			params: middleware.AcceptedVals{
				AccessToken: tokenKey,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			respBody:       returnVals{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockUserRepo := user.NewMockRepo()
			mockInterviewRepo := interview.NewMockRepo()
			mockTokenRepo := token.NewMockRepo()

			apiCfg := &apiConfig{
				UserRepo:      mockUserRepo,
				InterviewRepo: mockInterviewRepo,
				TokenRepo:     mockTokenRepo,
			}

			w, req, tc, err := setRequestAndWriter(http.MethodPost, "/api/auth/token", tc)
			if err != nil {
				t.Fatalf("failed to set request and writer")
			}

			req.Header.Set("Authorization", "Bearer "+tokenKey)

			// Apply the middleware to the handler
			handler := middleware.GetContext(http.HandlerFunc(apiCfg.refreshTokensHandler))

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			if w.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			// Validate resp
			resp, err := checkResponse(w, tc.respBody, tc.expectError)
			if err != nil {
				t.Fatalf("expected response %v and error %v\ngot response: %v and error %v", tc.respBody, tc.expectError, resp, resp.Error)
			}
		})
	}
}

func TestConversationsHandler_Post(t *testing.T) {
	tokenKey := "9942443a086328dfaa867e0708426f94284d25700fa9df930261e341f0d8c671"
	messagesResponse := ["Interviewer: What is a large donkey like?", "User: It's large!", "Interviewer: Yes, yes it is!"]
	tests := []TestCase{
		{
			name: "ConversationsHandler_Success",
			reqBody: `{"messages": [
				"Interviewer": "What is a large donkey like?",
				"User": "It's large!"
			]}`,
			params: middleware.AcceptedVals{
				AccessToken: tokenKey,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			respBody: returnVals{
				Messages: messagesResponse,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockUserRepo := user.NewMockRepo()
			mockInterviewRepo := interview.NewMockRepo()
			mockTokenRepo := token.NewMockRepo()

			apiCfg := &apiConfig{
				UserRepo:      mockUserRepo,
				InterviewRepo: mockInterviewRepo,
				TokenRepo:     mockTokenRepo,
			}

			w, req, tc, err := setRequestAndWriter(http.MethodPost, "/api/auth/token", tc)
			if err != nil {
				t.Fatalf("failed to set request and writer")
			}

			req.Header.Set("Authorization", "Bearer "+tokenKey)

			// Apply the middleware to the handler
			handler := middleware.GetContext(http.HandlerFunc(apiCfg.refreshTokensHandler))

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			if w.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			// Validate resp
			resp, err := checkResponse(w, tc.respBody, tc.expectError)
			if err != nil {
				t.Fatalf("expected response %v and error %v\ngot response: %v and error %v", tc.respBody, tc.expectError, resp, resp.Error)
			}
		})
	}
}

func setRequestAndWriter(method, endpoint string, tc TestCase) (*httptest.ResponseRecorder, *http.Request, TestCase, error) {
	req := httptest.NewRequest(method, endpoint, strings.NewReader(tc.reqBody))
	req = req.WithContext(context.WithValue(req.Context(), "params", tc.params))
	w := httptest.NewRecorder()

	return w, req, tc, nil
}

func checkResponse(w *httptest.ResponseRecorder, respBody returnVals, expectError bool) (returnVals, error) {
	var resp returnVals

	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		return resp, err
	}

	if !reflect.DeepEqual(resp, respBody) {
		return resp, err
	}

	if expectError && resp.Error == "" {
		return resp, err
	}
	return resp, nil
}

func createJWT(id, expires int) (string, error) {
	var (
		key   []byte
		token *jwt.Token
	)

	jwtSecret := os.Getenv("JWT_SECRET")
	now := time.Now()
	if expires == 0 {
		expires = 3600
	}
	expiresAt := time.Now().Add(time.Duration(expires) * time.Second)
	key = []byte(jwtSecret)
	claims := jwt.RegisteredClaims{
		Issuer:    "interviewer",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Subject:   strconv.Itoa(id),
	}
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(key)
	if err != nil {
		log.Fatalf("Bad SignedString: %s", err)
		return "", err
	}

	return signedToken, nil
}
