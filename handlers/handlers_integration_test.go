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
	respBodyFunc   func() handlers.ReturnVals
	DBCheck        bool
	Interview      *interview.Interview
	Conversation   *conversation.Conversation
	User           *user.User
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
	Handler = testutil.InitTestServer()

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

func Test_CreateUsersHandler_Integration(t *testing.T) {
	tests := []TestCase{
		{
			name:   "CreateUser_Success",
			method: "POST",
			url:    testutil.TestServerURL + "/api/users",
			reqBody: `{
				"username": "test",
				"email" : "test@test.com",
				"password" : "test"
			}`,
			expectedStatus: http.StatusCreated,
			respBody: handlers.ReturnVals{
				UserID:   1,
				Username: "test",
				Email:    "test@test.com",
			},
			DBCheck: true,
			User: &user.User{
				ID:       1,
				Username: "test",
				Email:    "test@test.com",
			},
		},
		{
			name:   "CreateUser_MissingUsername",
			method: "POST",
			url:    testutil.TestServerURL + "/api/users",
			reqBody: `{
				"email" : "test@test.com",
				"password" : "test"
			}`,
			expectedStatus: http.StatusBadRequest,
			respBody: handlers.ReturnVals{
				Error: "Username, Email, and Password required",
			},
		},
		{
			name:   "CreateUser_DuplicateUsername",
			method: "POST",
			url:    testutil.TestServerURL + "/api/users",
			reqBody: `{
				"username": "testUser",
				"email": "test@test.com",
				"password": "test"
			}`,
			expectedStatus: http.StatusConflict,
			respBody: handlers.ReturnVals{
				Error: "Email or username already exists",
			},
		},
		{
			name:   "CreateUser_MissingEmail",
			method: "POST",
			url:    testutil.TestServerURL + "/api/users",
			reqBody: `{
				"username" : "test",
				"password" : "test"
			}`,
			expectedStatus: http.StatusBadRequest,
			respBody: handlers.ReturnVals{
				Error: "Username, Email, and Password required",
			},
		},
		{
			name:   "CreateUser_DuplicateEmail",
			method: "POST",
			url:    testutil.TestServerURL + "/api/users",
			reqBody: `{
				"username": "test",
				"email": "testUser@test.com",
				"password": "test"
			}`,
			expectedStatus: http.StatusConflict,
			respBody: handlers.ReturnVals{
				Error: "Email or username already exists",
			},
		},
		{
			name:   "CreateUser_MissingPassword",
			method: "POST",
			url:    testutil.TestServerURL + "/api/users",
			reqBody: `{
				"username" : "test",
				"email": "test@test.com"
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
					t.Fatalf("Assert Database: GetUser failing: %v", err)
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
	t.Skip("TODO: refactor rest of tests to generate testuser/jwt in isolation. Current global dependency generation in Main is flaky")
	tests := []TestCase{
		{
			name:           "GetUser_Success",
			method:         "GET",
			url:            testutil.TestServerURL + "/api/users/2",
			expectedStatus: http.StatusOK,
			respBody: handlers.ReturnVals{
				UserID:   2,
				Username: "test",
				Email:    "test@test.com",
			},
			DBCheck: true,
			User: &user.User{
				ID:       2,
				Username: "test",
				Email:    "test@test.com",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp, respCode, err := testRequests(t, tc.headerKey, tc.headerValue, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				log.Fatalf("TestRequest for interview creation failed: %v", err)
			}

			//DEBUG
			fmt.Printf("\n\nresp: %s", resp)

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
					t.Fatalf("Assert Database: GetUser failing: %v", err)
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

func Test_InterviewsHandler_Integration(t *testing.T) {
	jwtoken, userID := testutil.CreateTestUserAndJWT()
	expiredJWT := testutil.CreateTestJWT(userID, -1)

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
			name:           "CreateInterview_MalformedHeaderValue",
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
					t.Fatalf("Assert Database: GetInterview failing: %v", err)
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
	jwtoken, userID := testutil.CreateTestUserAndJWT()
	expiredJWT := testutil.CreateTestJWT(userID, -1)

	tests := []TestCase{
		{
			name:   "CreateConversation_Success",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/create/1",
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
			name:   "CreateConversation_MissingBearer&Token",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/create/1",
			reqBody: `{
				"message" : "Answer1"
			}`,
			headerKey:      "Authorization",
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
			DBCheck: false,
		},
		{
			name:   "CreateConversation_MissingToken",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/create/1",
			reqBody: `{
				"message" : "Answer1"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
			DBCheck: false,
		},
		{
			name:   "CreateConversation_MalformedHeaderValue",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/create/1",
			reqBody: `{
				"message" : "Answer1"
			}`,
			headerKey:      "Authorization",
			headerValue:    "as9d8f7as09d87",
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
			DBCheck: false,
		},
		{
			name:   "CreateConversation_ExpiredToken",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/create/1",
			reqBody: `{
				"message" : "Answer1"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + expiredJWT,
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
			DBCheck: false,
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

			if diff := cmp.Diff(expected, got, cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Errorf("Mismatch (-expected +got):\n%s", diff)
			}

			// Assert Database
			if tc.DBCheck {
				conversation, err := conversation.GetConversation(Handler.ConversationRepo, got.Conversation.ID)
				if err != nil {
					t.Fatalf("Assert Database: GetConversation failing: %v", err)
				}

				expectedDB := expected.Conversation
				gotDB := conversation

				if diff := cmp.Diff(expectedDB, gotDB, cmpopts.EquateApproxTime(time.Second)); diff != "" {
					t.Errorf("Mismatch (-expected +got):\n%s", diff)
				}
			}
		})
	}
}

func Test_AppendConversationsHandler_Integration(t *testing.T) {
	jwtoken, userID := testutil.CreateTestUserAndJWT()
	expiredJWT := testutil.CreateTestJWT(userID, -1)

	tests := []TestCase{
		{
			name:   "AppendConversation_Success",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/append/1",
			reqBody: `{
				"conversation_id" : 1,
				"message" : "Answer2"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusCreated,
			respBodyFunc:   conversationBuilder.NewAppendedConversationMock(),
			DBCheck:        true,
		},
		{
			name:   "AppendConversation_MissingBearer&Token",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/append/1",
			reqBody: `{
				"conversation_id" : 1,
				"message" : "Answer2"
			}`,
			headerKey:      "Authorization",
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
			DBCheck: false,
		},
		{
			name:   "AppendConversation_MissingToken",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/append/1",
			reqBody: `{
				"conversation_id" : 1,
				"message" : "Answer2"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
			DBCheck: false,
		},
		{
			name:   "AppendConversation_MalformedHeaderValue",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/append/1",
			reqBody: `{
				"conversation_id" : 1,
				"message" : "Answer2"
			}`,
			headerKey:      "Authorization",
			headerValue:    "as9d8f7as09d87",
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
			DBCheck: false,
		},
		{
			name:   "AppendConversation_ExpiredToken",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/append/1",
			reqBody: `{
				"conversation_id" : 1,
				"message" : "Answer2"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + expiredJWT,
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
			DBCheck: false,
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
			url:    testutil.TestServerURL + "/api/conversations/append/1",
			reqBody: `{
				"conversation_id" : 1,
				"message" : "Answer1"
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusCreated,
			respBodyFunc:   conversationBuilder.NewIsFinishedConversationMock(),
			DBCheck:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

			if diff := cmp.Diff(expected, got, cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Errorf("Mismatch (-expected +got):\n%s", diff)
			}

			// Assert Database
			if tc.DBCheck {
				conversation, err := conversation.GetConversation(Handler.ConversationRepo, got.Conversation.ID)
				if err != nil {
					t.Fatalf("Assert Database: GetConversation failing: %v", err)
				}

				expectedDB := expected.Conversation
				gotDB := conversation

				if diff := cmp.Diff(expectedDB, gotDB, cmpopts.EquateApproxTime(time.Second)); diff != "" {
					t.Errorf("Mismatch (-expected +got):\n%s", diff)
				}
			}
		})
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
