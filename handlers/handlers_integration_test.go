package handlers_test

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/internal/testutil"
	"github.com/michaelboegner/interviewer/middleware"
)

func TestInterviewsHandler_Post_Integration(t *testing.T) {
	_, jwt, _ := testutil.CreateTestUserAndJWT(t)

	tests := []TestCase{
		{
			name:           "CreateInterview_Success",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/interviews",
			reqBody:        `{}`,
			headerType:     "Authorization",
			header:         "Bearer " + jwt,
			params:         middleware.AcceptedVals{},
			expectedStatus: http.StatusCreated,
			respBody: handlers.ReturnVals{
				InterviewID:   1,
				FirstQuestion: "Tell me a little bit about your work history.",
			},
		},
		{
			name:           "CreateInterview_MissingToken",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/interviews",
			reqBody:        `{}`,
			headerType:     "Authorization",
			header:         "",
			params:         middleware.AcceptedVals{},
			expectedStatus: http.StatusUnauthorized,
			respBody: handlers.ReturnVals{
				Error: "Unauthorized",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			interviewResp, respCode, err := testutil.TestRequests(t, tc.headerType, tc.header, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				log.Fatalf("TestRequest for interview creation failed: %v", err)
			}

			interviewsUnmarshalled := &handlers.ReturnVals{}
			err = json.Unmarshal(interviewResp, interviewsUnmarshalled)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Assert
			if respCode != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, respCode)
			}

			got := *interviewsUnmarshalled
			expected := tc.respBody

			if diff := cmp.Diff(expected, got); diff != "" {
				t.Errorf("Mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}
