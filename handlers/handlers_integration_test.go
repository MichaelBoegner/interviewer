package handlers_test

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/internal/mocks"
	"github.com/michaelboegner/interviewer/internal/testutil"
	"github.com/michaelboegner/interviewer/interview"
)

func Test_InterviewsHandler_Post_Integration(t *testing.T) {
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
				FirstQuestion: "Tell me a little bit about your work history.",
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
				FirstQuestion:   "Tell me a little bit about your work history.",
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

func Test_ConversationsHandler_Post_Integration(t *testing.T) {

	tests := []TestCase{
		{
			name:   "CreateConversation_Success",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/1",
			reqBody: `{
				"message" : "I have been a TSE for 5 years."
			}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwtoken,
			expectedStatus: http.StatusCreated,
			respBody: handlers.ReturnVals{
				Conversation: mocks.CreatedConversationMock,
			},
			DBCheck:      true,
			Conversation: mocks.CreatedConversationMock,
		},
		{
			name:   "CreateConversation_MissingBearer&Token",
			method: "POST",
			url:    testutil.TestServerURL + "/api/conversations/1",
			reqBody: `{
				"message" : "I have been a TSE for 5 years."
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
			url:    testutil.TestServerURL + "/api/conversations/1",
			reqBody: `{
				"message" : "I have been a TSE for 5 years."
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
			url:    testutil.TestServerURL + "/api/conversations/1",
			reqBody: `{
				"message" : "I have been a TSE for 5 years."
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
			url:    testutil.TestServerURL + "/api/conversations/1",
			reqBody: `{
				"message" : "I have been a TSE for 5 years."
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
			url:    testutil.TestServerURL + "/api/conversations/",
			reqBody: `{
				"message" : "I have been a TSE for 5 years."
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
			url:    testutil.TestServerURL + "/api/conversations/2",
			reqBody: `{
				"message" : "I have been a TSE for 5 years."
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

			expected := tc.respBody
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

				expectedDB := tc.Conversation
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
