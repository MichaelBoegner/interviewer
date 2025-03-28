package handlers_test

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/internal/testutil"
)

func TestInterviewsHandler_Post_Integration(t *testing.T) {
	_, jwt, userID := testutil.CreateTestUserAndJWT(t)
	expiredJWT, err := testutil.CreateTestJWT(t, userID, -1)
	if err != nil {
		log.Fatalf("CreateTestJWT failed to create expiredJWT: %v", err)
	}

	tests := []TestCase{
		{
			name:           "CreateInterview_Success",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/interviews",
			reqBody:        `{}`,
			headerKey:      "Authorization",
			headerValue:    "Bearer " + jwt,
			expectedStatus: http.StatusCreated,
			respBody: handlers.ReturnVals{
				InterviewID:   1,
				FirstQuestion: "Tell me a little bit about your work history.",
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

			// Assert
			if respCode != tc.expectedStatus {
				t.Fatalf("[%s]expected status %d, got %d\n", tc.name, tc.expectedStatus, respCode)
			}

			expected := tc.respBody
			got := *respUnmarshalled

			if diff := cmp.Diff(expected, got); diff != "" {
				t.Errorf("Mismatch (-expected +got):\n%s", diff)
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
