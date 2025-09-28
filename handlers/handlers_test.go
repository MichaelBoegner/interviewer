package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/joho/godotenv"
	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/internal/mocks"
	"github.com/michaelboegner/interviewer/internal/testutil"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

type TestCase struct {
	name           string
	username       string
	email          string
	password       string
	method         string
	url            string
	reqBody        string
	headerKey      string
	headerValue    string
	expectedStatus int
	respBody       handlers.ReturnVals
	respBodyFunc   func() handlers.ReturnVals
	Interview      *interview.Interview
	Conversation   *conversation.Conversation
	User           *user.User
	TokensExpected bool
	DBCheck        bool
	setup          func()
}

var (
	Handler             *handlers.Handler
	conversationBuilder *testutil.ConversationBuilder
	mockAI              *mocks.MockOpenAIClient
)

var logger *slog.Logger

func TestMain(m *testing.M) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger = slog.New(handler)

	logger.Info("Loading environment variables...")
	if err := godotenv.Load("../.env.test"); err != nil {
		logger.Error("failed to load .env.test", "error", err)
		os.Exit(1)
	}

	logger.Info("Initializing test server...")
	var err error
	Handler, err = testutil.InitTestServer(logger)
	if err != nil {
		logger.Error("test server initialization failed", "error", err)
		os.Exit(1)
	}

	if testutil.TestServerURL == "" {
		logger.Error("TestServerURL is empty! The server did not start properly")
		os.Exit(1)
	}

	logger.Info("Test server started", "url", testutil.TestServerURL)

	mockAI = Handler.OpenAI.(*mocks.MockOpenAIClient)
	conversationBuilder = testutil.NewConversationBuilder()

	code := m.Run()

	logger.Info("Stopping test server...")
	testutil.StopTestServer()

	os.Exit(code)
}

func Test_RequestVerificationHandler_Integration(t *testing.T) {
	cleanDBOrFail(t)

	tests := []TestCase{
		{
			name:   "Verification_Success",
			method: "POST",
			url:    testutil.TestServerURL + "/api/auth/request-verification",
			reqBody: `{
			"username":       "test",
			"email":          "test@test.com",
			"password":       "test"
			}`,
			expectedStatus: http.StatusOK,
			respBody: handlers.ReturnVals{
				Message: "Verification email sent",
			},
			DBCheck: false,
		},
		{
			name:   "CreateUser_MissingUsername",
			method: "POST",
			url:    testutil.TestServerURL + "/api/auth/request-verification",
			reqBody: `{
			"email":          "test1@test.com",
			"password":       "test1"
			}`,
			expectedStatus: http.StatusBadRequest,
			respBody: handlers.ReturnVals{
				Error: "Username, Email, and Password required",
			},
		},
		{
			name:   "CreateUser_MissingEmail",
			method: "POST",
			url:    testutil.TestServerURL + "/api/auth/request-verification",
			reqBody: `{
			"username":       "test1",
			"password":       "test1"
			}`,
			expectedStatus: http.StatusBadRequest,
			respBody: handlers.ReturnVals{
				Error: "Username, Email, and Password required",
			},
		},
		{
			name:   "CreateUser_MissingPassword",
			method: "POST",
			url:    testutil.TestServerURL + "/api/auth/request-verification",
			reqBody: `{
			"username":       "test1",
			"email":          "test1@test.com"
			}`,
			expectedStatus: http.StatusBadRequest,
			respBody: handlers.ReturnVals{
				Error: "Username, Email, and Password required",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				t.Fatalf("TestRequest failed: %v", err)
			}

			respUnmarshalled := &handlers.ReturnVals{}
			err = json.Unmarshal(resp, respUnmarshalled)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Assert Response
			if respCode != tc.expectedStatus {
				t.Fatalf("[%s] expected status %d, got %d\n", tc.name, tc.expectedStatus, respCode)
			}

			expected := tc.respBody
			got := *respUnmarshalled

			if diff := cmp.Diff(expected, got, cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Errorf("Mismatch (-expected +got):\n%s", diff)
			}

			// Assert Database
			if tc.DBCheck {
				user, err := user.GetUser(Handler.UserRepo, got.UserID)
				if err != nil {
					t.Fatalf("Assert Database: GetUser failed: %v", err)
				}

				expectedDB := tc.User
				gotDB := user

				if diff := cmp.Diff(expectedDB, gotDB, cmpopts.EquateApproxTime(time.Second)); diff != "" {
					t.Errorf("DB Mismatch (-expected +got):\n%s", diff)
				}
			}
		})
	}
}

func Test_CreateUsersHandler_Integration(t *testing.T) {
	cleanDBOrFail(t)

	tests := []TestCase{
		{
			name:           "CreateUser_Success",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/users",
			username:       "test",
			email:          "test@test.com",
			password:       "test",
			expectedStatus: http.StatusCreated,
			respBody: handlers.ReturnVals{
				UserID:   1,
				Username: "test",
				Email:    "test@test.com",
			},
			DBCheck: true,
			User: &user.User{
				ID:                 1,
				Username:           "test",
				Email:              "test@test.com",
				SubscriptionTier:   "free",
				SubscriptionStatus: "inactive",
				SubscriptionID:     "0",
				IndividualCredits:  1,
				AccountStatus:      "active",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			verificationJWT, err := user.VerificationToken(tc.email, tc.username, tc.password)
			if err != nil {
				t.Fatalf("GenerateEmailVerificationToken failed: %v", err)
			}
			reqBodyUser := strings.NewReader(fmt.Sprintf(`{
							"token": "%s"
							}`, verificationJWT))

			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, reqBodyUser)
			if err != nil {
				t.Fatalf("TestRequest failed: %v", err)
			}

			respUnmarshalled := &handlers.ReturnVals{}
			err = json.Unmarshal(resp, respUnmarshalled)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Assert Response
			if respCode != tc.expectedStatus {
				t.Fatalf("[%s] expected status %d, got %d\n", tc.name, tc.expectedStatus, respCode)
			}

			expected := tc.respBody
			got := *respUnmarshalled

			// Assert JWT is non-empty
			if got.JWToken == "" {
				t.Fatalf("[%s] expected JWT, got empty string", tc.name)
			}

			// Ignore JWToken when diffing
			got.JWToken = ""

			if diff := cmp.Diff(expected, got, cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Errorf("Mismatch (-expected +got):\n%s", diff)
			}

			// Assert Database
			if tc.DBCheck {
				user, err := user.GetUser(Handler.UserRepo, got.UserID)
				if err != nil {
					t.Fatalf("Assert Database: GetUser failed: %v", err)
				}

				expectedDB := tc.User
				gotDB := user

				if diff := cmp.Diff(expectedDB, gotDB, cmpopts.EquateApproxTime(time.Second)); diff != "" {
					t.Errorf("DB Mismatch (-expected +got):\n%s", diff)
				}
			}
		})
	}
}

func Test_GetUsersHandler_Integration(t *testing.T) {
	cleanDBOrFail(t)

	jwtoken, userID := testutil.CreateTestUserAndJWT(logger)

	tests := []TestCase{
		{
			name:           "GetUser_Success",
			method:         "GET",
			url:            testutil.TestServerURL + fmt.Sprintf("/api/users/%d", userID),
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusOK,
			respBody: handlers.ReturnVals{
				UserID:   userID,
				Username: "test",
				Email:    "test@test.com",
			},
			DBCheck: false,
		},
		{
			name:           "GetUser_IncorrectID",
			method:         "GET",
			url:            testutil.TestServerURL + "/api/users/2",
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Invalid ID",
			},
			DBCheck: false,
		},
		{
			name:           "GetUser_MissingID",
			method:         "GET",
			url:            testutil.TestServerURL + "/api/users/",
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusBadRequest,
			respBody: handlers.ReturnVals{
				Error: "UserID required",
			},
			DBCheck: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				t.Fatalf("TestRequest failed: %v", err)
			}

			respUnmarshalled := &handlers.ReturnVals{}
			err = json.Unmarshal(resp, respUnmarshalled)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Assert Response
			if respCode != tc.expectedStatus {
				t.Fatalf("[%s] expected status %d, got %d\n", tc.name, tc.expectedStatus, respCode)
			}

			expected := tc.respBody
			got := *respUnmarshalled

			if diff := cmp.Diff(expected, got, cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Errorf("Mismatch (-expected +got):\n%s", diff)
			}

			// Assert Database
			if tc.DBCheck {
				user, err := user.GetUser(Handler.UserRepo, got.UserID)
				if err != nil {
					t.Fatalf("Assert Database: GetUser failed: %v", err)
				}

				expectedDB := tc.User
				gotDB := user

				if diff := cmp.Diff(expectedDB, gotDB, cmpopts.EquateApproxTime(time.Second)); diff != "" {
					t.Errorf("DB Mismatch (-expected +got):\n%s", diff)
				}
			}
		})
	}
}

func Test_LoginHandler_Integration(t *testing.T) {
	cleanDBOrFail(t)

	_, _ = testutil.CreateTestUserAndJWT(logger)

	tests := []TestCase{
		{
			name:           "Login_Success",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/auth/login",
			expectedStatus: http.StatusOK,
			reqBody: `{
				"email" : "test@test.com",
				"password" : "test"
			}`,
			DBCheck:        true,
			TokensExpected: true,
		},
		{
			name:           "Login_MissingUsername",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/auth/login",
			expectedStatus: http.StatusBadRequest,
			reqBody: `{
				"password": "test"
			}`,
			DBCheck:        false,
			TokensExpected: false,
		},
		{
			name:           "Login_MissingPassword",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/auth/login",
			expectedStatus: http.StatusBadRequest,
			reqBody: `{
				"username": "test"
			}`,
			DBCheck:        false,
			TokensExpected: false,
		},
		{
			name:           "Login_WrongEmail",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/auth/login",
			expectedStatus: http.StatusUnauthorized,
			reqBody: `{
				"email": "notarealemail@test.com",
				"password": "test"
			}`,
			DBCheck:        false,
			TokensExpected: false,
		},
		{
			name:           "Login_WrongPassword",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/auth/login",
			expectedStatus: http.StatusUnauthorized,
			reqBody: `{
				"email": "test@test.com",
				"password": "wrongpass"
			}`,
			DBCheck:        false,
			TokensExpected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				t.Fatalf("TestRequest failed: %v", err)
			}

			respUnmarshalled := &handlers.ReturnVals{}
			err = json.Unmarshal(resp, respUnmarshalled)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Assert Response
			if respCode != tc.expectedStatus {
				t.Fatalf("[%s] expected status %d, got %d\n", tc.name, tc.expectedStatus, respCode)
			}

			if tc.TokensExpected {
				if respUnmarshalled.JWToken == "" {
					t.Fatalf("Expected access token, got empty string")
				}

				if respUnmarshalled.RefreshToken == "" {
					t.Fatalf("Expected refresh token, got empty string")
				}
			} else {
				if respUnmarshalled.JWToken != "" {
					t.Fatalf("Did not expect JWT, but got one: %v", respUnmarshalled.JWToken)
				}

				if respUnmarshalled.RefreshToken != "" {
					t.Fatalf("Did not expect refresh token, but got one: %v", respUnmarshalled.RefreshToken)
				}
			}

			// Assert Database
			if tc.DBCheck {
				refreshToken, err := token.GetStoredRefreshToken(Handler.TokenRepo, respUnmarshalled.UserID)
				if err != nil {
					t.Fatalf("Assert Database: GetUser failed: %v", err)
				}

				expectedDB := respUnmarshalled.RefreshToken
				gotDB := refreshToken

				if diff := cmp.Diff(expectedDB, gotDB, cmpopts.EquateApproxTime(time.Second)); diff != "" {
					t.Errorf("DB Mismatch (-expected +got):\n%s", diff)
				}
			}
		})
	}
}

func Test_RefreshTokensHandler_Integration(t *testing.T) {
	cleanDBOrFail(t)

	_, userID := testutil.CreateTestUserAndJWT(logger)
	refreshToken, err := token.GetStoredRefreshToken(Handler.TokenRepo, userID)
	if err != nil {
		t.Fatalf("TC GetStoredRefreshToken failed: %v", err)
	}

	tests := []TestCase{
		{
			name:           "RefreshToken_Success",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/auth/token",
			expectedStatus: http.StatusOK,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + refreshToken,
			reqBody: fmt.Sprintf(`{
				"user_id" : %d
			}`, userID),
			DBCheck:        true,
			TokensExpected: true,
		},
		{
			name:           "RefreshToken_IncorrectUserID",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/auth/token",
			expectedStatus: http.StatusBadRequest,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + refreshToken,
			reqBody: `{
				"user_id" : 2
			}`,
			DBCheck:        false,
			TokensExpected: false,
		},
		{
			name:           "RefreshToken_MissingBearer&Token",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/auth/token",
			expectedStatus: http.StatusUnauthorized,
			headerKey:      "Authorization",
			reqBody: fmt.Sprintf(`{
				"user_id": %d
			}`, userID),
			DBCheck:        false,
			TokensExpected: false,
		},
		{
			name:           "RefreshToken_MissingToken",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/auth/token",
			expectedStatus: http.StatusUnauthorized,
			headerKey:      "Authorization",
			headerValue:    "Bearer ",
			reqBody: fmt.Sprintf(`{
				"user_id": %d
			}`, userID),
			DBCheck:        false,
			TokensExpected: false,
		},
		{
			name:           "RefreshToken_MalformedToken",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/auth/token",
			expectedStatus: http.StatusUnauthorized,
			headerKey:      "Authorization",
			headerValue:    "12341234",
			reqBody: fmt.Sprintf(`{
				"user_id": %d
			}`, userID),
			DBCheck:        false,
			TokensExpected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				t.Fatalf("TestRequest failed: %v", err)
			}

			respUnmarshalled := &handlers.ReturnVals{}
			err = json.Unmarshal(resp, respUnmarshalled)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Assert Response
			if respCode != tc.expectedStatus {
				t.Fatalf("[%s] expected status %d, got %d\n", tc.name, tc.expectedStatus, respCode)
			}

			if tc.TokensExpected {
				if respUnmarshalled.JWToken == "" {
					t.Fatalf("Expected access token, got empty string")
				}

				if respUnmarshalled.RefreshToken == "" {
					t.Fatalf("Expected refresh token, got empty string")
				}
			} else {
				if respUnmarshalled.JWToken != "" {
					t.Fatalf("Did not expect JWT, but got one: %v", respUnmarshalled.JWToken)
				}

				if respUnmarshalled.RefreshToken != "" {
					t.Fatalf("Did not expect refresh token, but got one: %v", respUnmarshalled.RefreshToken)
				}
			}

			// Assert Database
			if tc.DBCheck {
				refreshToken, err := token.GetStoredRefreshToken(Handler.TokenRepo, userID)
				if err != nil {
					t.Fatalf("Assert Database: GetUser failed: %v", err)
				}

				expectedDB := respUnmarshalled.RefreshToken
				gotDB := refreshToken

				if diff := cmp.Diff(expectedDB, gotDB, cmpopts.EquateApproxTime(time.Second)); diff != "" {
					t.Errorf("DB Mismatch (-expected +got):\n%s", diff)
				}
			}
		})
	}
}

func Test_InterviewsHandler_Integration(t *testing.T) {
	cleanDBOrFail(t)

	jwtoken, userID := testutil.CreateTestUserAndJWT(logger)
	expiredJWT := testutil.CreateTestExpiredJWT(userID, -1, logger)

	tests := []TestCase{
		{
			name:           "CreateInterview_Success",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/interviews",
			reqBody:        `{}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusCreated,
			respBody: handlers.ReturnVals{
				InterviewID:    1,
				ConversationID: 1,
				FirstQuestion:  "Question1",
			},
			DBCheck: true,
			Interview: &interview.Interview{
				Id:              1,
				ConversationID:  1,
				UserId:          1,
				Length:          30,
				NumberQuestions: 3,
				Difficulty:      "easy",
				Status:          "active",
				Score:           100,
				Language:        "Python",
				Prompt:          mocks.BuildTestPrompt([]string{}, "Introduction", 1, ""),
				FirstQuestion:   "Question1",
				Subtopic:        "None",
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			},
			setup: func() {
				mockAI.Scenario = mocks.ScenarioInterview
			},
		},
		{
			name:           "CreateInterview_MissingBearer&Token",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/interviews",
			reqBody:        `{}`,
			headerKey:      "Authorization",
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
		},
		{
			name:           "CreateInterview_MissingToken",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/interviews",
			reqBody:        `{}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
		},
		{
			name:           "CreateInterview_MalformedToken",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/interviews",
			reqBody:        `{}`,
			headerKey:      "Authorization",
			headerValue:    "as9d8f7as09d87",
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
		},
		{
			name:           "CreateInterview_ExpiredToken",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/interviews",
			reqBody:        `{}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + expiredJWT,
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				t.Fatalf("TestRequest failed: %v", err)
			}

			respUnmarshalled := &handlers.ReturnVals{}
			err = json.Unmarshal(resp, respUnmarshalled)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Assert Response
			if respCode != tc.expectedStatus {
				t.Fatalf("[%s] expected status %d, got %d\n", tc.name, tc.expectedStatus, respCode)
			}

			expected := tc.respBody
			got := *respUnmarshalled

			if diff := cmp.Diff(expected, got); diff != "" {
				t.Errorf("Mismatch (-expected +got):\n%s", diff)
			}

			// Assert Database
			if tc.DBCheck {
				interviewReturned, err := interview.GetInterview(Handler.InterviewRepo, respUnmarshalled.InterviewID)
				if err != nil {
					t.Fatalf("Assert Database: GetInterview failed: %v", err)
				}

				expectedDB := tc.Interview
				gotDB := interviewReturned

				if diff := cmp.Diff(expectedDB, gotDB, cmpopts.IgnoreFields(interview.Interview{}, "CreatedAt", "UpdatedAt")); diff != "" {
					t.Errorf("Mismatch (-expected +got):\n%s", diff)
				}
			}
		})
	}
}

func Test_CreateConversationsHandler_Integration(t *testing.T) {
	cleanDBOrFail(t)

	jwtoken, _ := testutil.CreateTestUserAndJWT(logger)
	mockAI.Scenario = mocks.ScenarioInterview
	interviewID := testutil.CreateTestInterview(jwtoken, logger)
	conversationsURL := testutil.TestServerURL + fmt.Sprintf("/api/conversations/create/%d", interviewID)

	tests := []TestCase{
		{
			name:   "CreateConversation_Success",
			method: "POST",
			url:    conversationsURL,
			reqBody: `{
				"message" : "T1Q1A1"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusCreated,
			respBodyFunc:   conversationBuilder.NewCreatedConversationMock(),
			DBCheck:        true,
			setup: func() {
				mockAI.Scenario = mocks.ScenarioCreated
			},
		},
		{
			name:   "CreateConversation_MissingIntervewID",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/create/",
			reqBody: `{
				"message" : "T1Q1A1"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusBadRequest,
			respBody: handlers.ReturnVals{
				Error: "Missing ID",
			},
			DBCheck: false,
		},
		{
			name:   "CreateConversation_IncorrectInterviewID",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/create/2",
			reqBody: `{
				"message" : "T1Q1A1"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusBadRequest,
			respBody: handlers.ReturnVals{
				Error: "Invalid ID",
			},
			DBCheck: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				t.Fatalf("TestRequest failed: %v", err)
			}

			respUnmarshalled := &handlers.ReturnVals{}
			err = json.Unmarshal(resp, respUnmarshalled)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Assert Response
			if respCode != tc.expectedStatus {
				t.Fatalf("[%s] expected status %d, got %d\n", tc.name, tc.expectedStatus, respCode)
			}

			var expected handlers.ReturnVals

			if tc.respBodyFunc != nil {
				expected = tc.respBodyFunc()
			} else {
				expected = tc.respBody
			}
			got := *respUnmarshalled

			if diff := cmp.Diff(expected, got, cmpopts.EquateApproxTime(3*time.Second)); diff != "" {
				t.Errorf("Mismatch (-expected +got):\n%s", diff)
			}

			// Assert Database
			if tc.DBCheck {
				conversation, err := conversation.GetConversation(Handler.ConversationRepo, got.Conversation.ID)
				if err != nil {
					t.Fatalf("Assert Database: GetConversation failed: %v", err)
				}

				expectedDB := expected.Conversation
				gotDB := conversation

				if diff := cmp.Diff(expectedDB, gotDB, cmpopts.EquateApproxTime(3*time.Second)); diff != "" {
					t.Errorf("Mismatch (-expected +got):\n%s", diff)
				}
			}
		})
	}
}

func Test_AppendConversationsHandler_Integration(t *testing.T) {
	cleanDBOrFail(t)
	jwtoken, _ := testutil.CreateTestUserAndJWT(logger)
	mockAI.Scenario = mocks.ScenarioInterview

	interviewID := testutil.CreateTestInterview(jwtoken, logger)
	mockAI.Scenario = mocks.ScenarioCreated

	conversationID := testutil.CreateTestConversation(jwtoken, interviewID, logger)
	urlTest := testutil.TestServerURL + fmt.Sprintf("/api/conversations/append/%d", interviewID)

	tests := []TestCase{
		{
			name:   "AppendConversation_Success",
			method: "POST",
			url:    urlTest,
			reqBody: fmt.Sprintf(`{
				"conversation_id" : %d,
				"message" : "T1Q2A2"
			}`, conversationID),
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusCreated,
			respBodyFunc:   testutil.NewAppendedConversationMock(),
			DBCheck:        true,
			setup: func() {
				mockAI.Scenario = mocks.ScenarioAppended1
			},
		},
		{
			name:   "AppendConversation_MissingInterviewID",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/append/",
			reqBody: `{
				"conversation_id" : 1,
				"message" : "Answer2"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusBadRequest,
			respBody: handlers.ReturnVals{
				Error: "Missing ID",
			},
			DBCheck: false,
		},
		{
			name:   "AppendConversation_IncorrectInterviewID",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/append/9999",
			reqBody: `{
				"conversation_id" : 1,
				"message" : "Answer2"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusBadRequest,
			respBody: handlers.ReturnVals{
				Error: "Invalid ID",
			},
			DBCheck: false,
		},
		{
			name:   "AppendConversation_IsFinished",
			method: "POST",
			url:    urlTest,
			reqBody: fmt.Sprintf(`{
				"conversation_id" : %d,
				"message" : "T2Q2A2"
			}`, conversationID),
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusCreated,
			respBodyFunc:   testutil.NewIsFinishedConversationMock(),
			DBCheck:        true,
			setup: func() {
				mockAI.Scenario = mocks.ScenarioIsFinished
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "AppendConversation_IsFinished" {
				reqBodyPre := fmt.Sprintf(`{
					"conversation_id" : %d,
					"message" : "T2Q1A1"
				}`, conversationID)
				mockAI.Scenario = mocks.ScenarioAppended2
				_, _, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(reqBodyPre))
				if err != nil {
					t.Fatalf("Precondition request failed: %v", err)
				}
			}

			if tc.setup != nil {
				tc.setup()
			}

			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				t.Fatalf("TestRequest failed: %v", err)
			}

			respUnmarshalled := &handlers.ReturnVals{}
			if err := json.Unmarshal(resp, respUnmarshalled); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if respCode != tc.expectedStatus {
				t.Fatalf("[%s] expected status %d, got %d", tc.name, tc.expectedStatus, respCode)
			}

			expected := tc.respBody
			if tc.respBodyFunc != nil {
				expected = tc.respBodyFunc()
			}

			if diff := cmp.Diff(expected, *respUnmarshalled, cmpopts.EquateApproxTime(3*time.Second)); diff != "" {
				t.Errorf("Mismatch (-expected +got):\n%s", diff)
			}

			// DB validation
			if tc.DBCheck {
				gotDB, err := conversation.GetConversation(Handler.ConversationRepo, respUnmarshalled.Conversation.ID)
				if err != nil {
					t.Fatalf("DB check failed: %v", err)
				}

				if diff := cmp.Diff(expected.Conversation, gotDB, cmpopts.EquateApproxTime(3*time.Second)); diff != "" {
					t.Errorf("DB mismatch (-expected +got):\n%s", diff)
				}
			}
		})
	}
}

func cleanDBOrFail(t *testing.T) {
	if err := testutil.TruncateAllTables(Handler.DB); err != nil {
		t.Fatalf("Failed to clean DB: %v", err)
	}
}

func testRequests(t *testing.T, headerKey, headerValue, method, url string, reqBody *strings.Reader) ([]byte, int, error) {
	t.Helper()
	client := &http.Client{}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Logf("CreateTestUserAndJWT user creation failed: %v", err)
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if headerKey != "" {
		req.Header.Set(headerKey, headerValue)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Request failed: %v", err)
		return nil, resp.StatusCode, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Reading response failed: %v", err)
		return nil, resp.StatusCode, err
	}

	return bodyBytes, resp.StatusCode, nil
}
