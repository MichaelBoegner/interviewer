package handlers_test

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
	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/internal/testutil"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

type TestCase struct {
	name           string
	method         string
	url            string
	reqBody        string
	headerKey      string
	headerValue    string
	params         middleware.AcceptedVals
	expectedStatus int
	expectError    bool
	respBody       handlers.ReturnVals
}

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Llongfile)

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

func TestUsersHandler_Post(t *testing.T) {
	t.Skip("TODO: Skipping until integration tests are in place or mocking is added")
	tests := []TestCase{
		{
			name:    "CreateUser_Success",
			reqBody: `{"username":"testuser", "email":"test@example.com", "password":"password"}`,
			params: middleware.AcceptedVals{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
			respBody: handlers.ReturnVals{
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
			respBody:       handlers.ReturnVals{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockUserRepo := user.NewMockRepo()
			handler := &handlers.Handler{UserRepo: mockUserRepo}

			w, req := setRequestAndWriter(http.MethodPost, "/api/users", tc)

			// Act
			handler.UsersHandler(w, req)

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
	t.Skip("TODO: Skipping until integration tests are in place or mocking is added")
	tests := []TestCase{
		{
			name:           "GetUser_Success",
			reqBody:        `{}`,
			params:         middleware.AcceptedVals{},
			expectedStatus: http.StatusOK,
			expectError:    false,
			respBody: handlers.ReturnVals{
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
			handler := &handlers.Handler{UserRepo: mockUserRepo}

			w, req := setRequestAndWriter(http.MethodGet, "/api/users/1", tc)

			// Act
			handler.UsersHandler(w, req)

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
	t.Skip("TODO: Skipping until integration tests are in place or mocking is added")
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
			respBody: handlers.ReturnVals{
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
			respBody:       handlers.ReturnVals{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockTokenRepo := token.NewMockRepo()
			mockUserRepo := user.NewMockRepo()
			handler := &handlers.Handler{
				TokenRepo: mockTokenRepo,
				UserRepo:  mockUserRepo,
			}

			w, req := setRequestAndWriter(http.MethodPost, "/api/auth/login", tc)

			// Act
			handler.LoginHandler(w, req)

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
	t.Skip("TODO: Skipping until integration tests are in place or mocking is added")
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
			respBody:       handlers.ReturnVals{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockUserRepo := user.NewMockRepo()
			mockInterviewRepo := interview.NewMockRepo()

			handler := &handlers.Handler{
				UserRepo:      mockUserRepo,
				InterviewRepo: mockInterviewRepo,
			}

			w, req := setRequestAndWriter(http.MethodPost, "/api/interviews", tc)

			req.Header.Set("Authorization", "Bearer "+token)

			// Apply the middleware to the handler
			handlerWithMiddleware := middleware.GetContext(http.HandlerFunc(handler.InterviewsHandler))

			// Act
			handlerWithMiddleware.ServeHTTP(w, req)

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
	t.Skip("TODO: Skipping until integration tests are in place or mocking is added")
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
			respBody:       handlers.ReturnVals{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockUserRepo := user.NewMockRepo()
			mockInterviewRepo := interview.NewMockRepo()
			mockTokenRepo := token.NewMockRepo()

			handler := &handlers.Handler{
				UserRepo:      mockUserRepo,
				InterviewRepo: mockInterviewRepo,
				TokenRepo:     mockTokenRepo,
			}

			w, req := setRequestAndWriter(http.MethodPost, "/api/auth/token", tc)

			req.Header.Set("Authorization", "Bearer "+tokenKey)

			// Apply the middleware to the handler
			handlerWithMiddleware := middleware.GetContext(http.HandlerFunc(handler.RefreshTokensHandler))

			// Act
			handlerWithMiddleware.ServeHTTP(w, req)

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
	t.Skip("TODO: Skipping until integration tests are in place or mocking is added")

	tokenKey := "9942443a086328dfaa867e0708426f94284d25700fa9df930261e341f0d8c671"
	conversationResponse := &conversation.Conversation{
		ID:          1,
		InterviewID: 1,
		Topics:      conversation.PredefinedTopics,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	topic := conversationResponse.Topics[1]
	topic.ConversationID = 1
	topic.Questions = make(map[int]*conversation.Question)
	question := &conversation.Question{
		ConversationID: 1,
		QuestionNumber: 1,
		Prompt:         "What is the flight speed of an unladdened swallow?",
	}

	messageFirst := &conversation.Message{
		ConversationID: 1,
		QuestionNumber: 1,
		Author:         "interviewer",
		Content:        "What is the flight speed of an unladdened swallow?",
		CreatedAt:      time.Now(),
	}

	messageResponse := &conversation.Message{
		ConversationID: 1,
		QuestionNumber: 1,
		Author:         "user",
		Content:        "European or African?",
		CreatedAt:      time.Now(),
	}

	messageResponse.ConversationID = 1
	messageResponse.QuestionNumber = 1
	messageResponse.CreatedAt = time.Now()

	question.Messages = make([]conversation.Message, 0)
	question.Messages = append(question.Messages, *messageFirst)
	question.Messages = append(question.Messages, *messageResponse)

	conversationResponse.Topics[1] = topic
	conversationResponse.Topics[1].Questions[1] = question

	tests := []TestCase{
		{
			name: "ConversationsHandler_Success",
			reqBody: `{"message": {
				"author": "user",
				"content": "European or African?"
			}}`,
			params: middleware.AcceptedVals{
				AccessToken: tokenKey,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			respBody: handlers.ReturnVals{
				Conversation: conversationResponse,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockUserRepo := user.NewMockRepo()
			mockInterviewRepo := interview.NewMockRepo()
			mockTokenRepo := token.NewMockRepo()
			mockConversationRepo := conversation.NewMockRepo()

			handler := &handlers.Handler{
				UserRepo:         mockUserRepo,
				InterviewRepo:    mockInterviewRepo,
				TokenRepo:        mockTokenRepo,
				ConversationRepo: mockConversationRepo,
			}

			w, req := setRequestAndWriter(http.MethodPost, "/api/conversations/1", tc)

			req.Header.Set("Authorization", "Bearer "+tokenKey)

			// Apply the middleware to the handler
			handlerWithMiddleware := middleware.GetContext(http.HandlerFunc(handler.ConversationsHandler))

			// Act
			handlerWithMiddleware.ServeHTTP(w, req)

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

func setRequestAndWriter(method, endpoint string, tc TestCase) (*httptest.ResponseRecorder, *http.Request) {
	req := httptest.NewRequest(method, endpoint, strings.NewReader(tc.reqBody))
	req = req.WithContext(context.WithValue(req.Context(), "params", tc.params))
	req.Header.Set("Content-Type", "application/json")
	if tc.headerKey != "" {
		req.Header.Set(tc.headerKey, tc.headerValue)
	}
	w := httptest.NewRecorder()

	return w, req
}

func checkResponse(w *httptest.ResponseRecorder, respBody handlers.ReturnVals, expectError bool) (handlers.ReturnVals, error) {
	var resp handlers.ReturnVals

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
