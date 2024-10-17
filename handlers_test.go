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

func TestMain(m *testing.M) {
	// Set the logging flags globally for all tests
	log.SetFlags(log.LstdFlags | log.Llongfile)

	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Run the tests
	code := m.Run()

	// Exit with the test suite's exit code
	os.Exit(code)
}

func TestUsersHandler_Post(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        string
		params         middleware.AcceptedVals
		expectedStatus int
		expectError    bool
		respBody       returnVals
	}{
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

			req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(tc.reqBody))
			req = req.WithContext(context.WithValue(req.Context(), "params", tc.params))
			w := httptest.NewRecorder()

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
	tests := []struct {
		name           string
		reqBody        string
		params         middleware.AcceptedVals
		expectedStatus int
		expectError    bool
		respBody       returnVals
	}{
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

			req := httptest.NewRequest(http.MethodGet, "/api/users/1", strings.NewReader(tc.reqBody))
			req = req.WithContext(context.WithValue(req.Context(), "params", tc.params))
			w := httptest.NewRecorder()

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

	tests := []struct {
		name           string
		reqBody        string
		params         middleware.AcceptedVals
		expectedStatus int
		expectError    bool
		respBody       returnVals
	}{
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

			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(tc.reqBody))
			req = req.WithContext(context.WithValue(req.Context(), "params", tc.params))
			w := httptest.NewRecorder()

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

	tests := []struct {
		name           string
		reqBody        string
		params         middleware.AcceptedVals
		expectedStatus int
		expectError    bool
		respBody       returnVals
	}{
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

			req := httptest.NewRequest(http.MethodPost, "/api/interviews", strings.NewReader(tc.reqBody))
			req = req.WithContext(context.WithValue(req.Context(), "params", tc.params))
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()

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
		key []byte
		t   *jwt.Token
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
	t = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(key)
	if err != nil {
		log.Fatalf("Bad SignedString: %s", err)
		return "", err
	}

	return s, nil
}
