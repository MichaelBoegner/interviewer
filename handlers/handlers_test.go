package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	"github.com/michaelboegner/interviewer/middleware"
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
	params         middleware.AcceptedVals
	expectedStatus int
	expectError    bool
	respBody       handlers.ReturnVals
	respBodyFunc   func() handlers.ReturnVals
	Interview      *interview.Interview
	Conversation   *conversation.Conversation
	User           *user.User
	TokensExpected bool
	DBCheck        bool
}

var (
	Handler             *handlers.Handler
	conversationBuilder *testutil.ConversationBuilder
)

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	log.Println("Loading environment variables...")
	err := godotenv.Load("../.env.test")
	if err != nil {
		log.Fatalf("Error loading .env.test file: %v", err)
	}

	log.Println("Initializing test server...")
	Handler, err = testutil.InitTestServer()
	if err != nil {
		log.Fatalf("Test server initialization failed: %v", err)
	}

	if testutil.TestServerURL == "" {
		log.Fatal("TestMain: TestServerURL is empty! The server did not start properly.")
	}

	log.Printf("TestMain: Test server started successfully at: %s", testutil.TestServerURL)

	conversationBuilder = testutil.NewConversationBuilder()

	code := m.Run()

	log.Println("Stopping test server...")
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
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				log.Fatalf("TestRequest for interview creation failed: %v", err)
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
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			verificationJWT, err := user.VerificationToken(tc.email, tc.username, tc.password)
			if err != nil {
				log.Printf("GenerateEmailVerificationToken failed: %v", err)
			}
			reqBodyUser := strings.NewReader(fmt.Sprintf(`{
							"token": "%s"
							}`, verificationJWT))

			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, reqBodyUser)
			if err != nil {
				log.Fatalf("TestRequest for interview creation failed: %v", err)
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

func Test_GetUsersHandler_Integration(t *testing.T) {
	cleanDBOrFail(t)

	jwtoken, userID := testutil.CreateTestUserAndJWT()

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
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				log.Fatalf("TestRequest for interview creation failed: %v", err)
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

	_, _ = testutil.CreateTestUserAndJWT()

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
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				log.Fatalf("TestRequest for interview creation failed: %v", err)
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

	_, userID := testutil.CreateTestUserAndJWT()
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
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				log.Fatalf("TestRequest for interview creation failed: %v", err)
			}
			// DEBUG
			fmt.Printf("resp: %v\n\n\n", string(resp))

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
	t.Skip()
	cleanDBOrFail(t)

	jwtoken, userID := testutil.CreateTestUserAndJWT()
	expiredJWT := testutil.CreateTestExpiredJWT(userID, -1)

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
				InterviewID:   1,
				FirstQuestion: "Question1",
			},
			DBCheck: true,
			Interview: &interview.Interview{
				Id:              1,
				UserId:          1,
				Length:          30,
				NumberQuestions: 3,
				Difficulty:      "easy",
				Status:          "Running",
				Score:           100,
				Language:        "Python",
				Prompt:          mocks.TestPrompt,
				FirstQuestion:   "Question1",
				Subtopic:        "None",
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
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				log.Fatalf("TestRequest for interview creation failed: %v", err)
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
				interview, err := interview.GetInterview(Handler.InterviewRepo, respUnmarshalled.InterviewID)
				if err != nil {
					t.Fatalf("Assert Database: GetInterview failed: %v", err)
				}

				expectedDB := tc.Interview
				gotDB := interview

				if diff := cmp.Diff(expectedDB, gotDB); diff != "" {
					t.Errorf("Mismatch (-expected +got):\n%s", diff)
				}
			}
		})
	}
}

func Test_CreateConversationsHandler_Integration(t *testing.T) {
	t.Skip()
	cleanDBOrFail(t)

	jwtoken, _ := testutil.CreateTestUserAndJWT()
	interviewID := testutil.CreateTestInterview(jwtoken)
	conversationsURL := testutil.TestServerURL + fmt.Sprintf("/api/conversations/create/%d", interviewID)

	tests := []TestCase{
		{
			name:   "CreateConversation_Success",
			method: "POST",
			url:    conversationsURL,
			reqBody: `{
				"message" : "Answer1"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusCreated,
			respBodyFunc:   conversationBuilder.NewCreatedConversationMock(),
			DBCheck:        true,
		},
		{
			name:   "CreateConversation_MissingIntervewID",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/create/",
			reqBody: `{
				"message" : "Answer1"
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
				"message" : "Answer1"
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
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				log.Fatalf("TestRequest for interview creation failed: %v", err)
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
	t.Skip()
	cleanDBOrFail(t)

	jwtoken, _ := testutil.CreateTestUserAndJWT()
	interviewID := testutil.CreateTestInterview(jwtoken)
	conversationID := testutil.CreateTestConversation(jwtoken, interviewID)
	urlTest := testutil.TestServerURL + fmt.Sprintf("/api/conversations/append/%d", interviewID)
	reqBodyTest := fmt.Sprintf(`{
				"conversation_id" : %d,
				"message" : "Answer2"
			}`, conversationID)

	tests := []TestCase{
		{
			name:           "AppendConversation_Success",
			method:         "POST",
			url:            urlTest,
			reqBody:        reqBodyTest,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusCreated,
			respBodyFunc:   conversationBuilder.NewAppendedConversationMock(),
			DBCheck:        true,
		},
		{
			name:   "AppendConversation_MissingIntervewID",
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
			url:    testutil.TestServerURL + "/api/conversations/append/2",
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
			name:   "AppendConversation_isFinished",
			method: "POST",
			url:    urlTest,
			reqBody: `{
				"conversation_id" : 1,
				"message" : "Answer1"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusCreated,
			respBodyFunc:   conversationBuilder.NewIsFinishedConversationMock(),
			DBCheck:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				log.Fatalf("TestRequest for interview creation failed: %v", err)
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

func showLogsIfFail(t *testing.T, name string, buf strings.Builder) {
	log.SetOutput(os.Stderr)
	if t.Failed() {
		fmt.Printf("---- logs for test: %s ----\n%s\n", name, buf.String())
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
