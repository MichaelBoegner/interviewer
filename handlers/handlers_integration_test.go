package handlers_test

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/internal/testutil"
	"github.com/michaelboegner/interviewer/middleware"
)

func TestInterviewsHandler_Post_Integration(t *testing.T) {
	// t.Skip("TODO: skipping for now while setting up recorded demo")

	_, jwt, _ := testutil.CreateTestUserAndJWT(t)

	tests := []TestCase{
		{
			name:           "CreateInterview_Success",
			method:         "POST",
			url:            testutil.TestServerURL + "/api/interviews",
			reqBody:        `{}`,
			headerType:     "Authorization",
			header:         "Bearer:" + jwt,
			params:         middleware.AcceptedVals{},
			expectedStatus: http.StatusCreated,
			expectError:    false,
			respBody: handlers.ReturnVals{
				ID:            1,
				FirstQuestion: "Tell me about your background experience in general.",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			// type InterviewsReturn struct {
			// 	ID            int
			// 	FirstQuestion string
			// }
			// Act
			interviewResp, respCode, err := testutil.TestRequests(t, tc.headerType, tc.header, tc.method, tc.url, strings.NewReader(tc.reqBody))
			if err != nil {
				log.Fatalf("TestRequest for interview creation failed: %v", err)
			}

			interviewsUnmarshalled := &handlers.ReturnVals{}
			json.Unmarshal(interviewResp, interviewsUnmarshalled)

			// Assert
			if respCode != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, respCode)
			}

			// Validate resp
			resp, err := checkResponseIntegrations(interviewsUnmarshalled, tc.respBody)
			if err != nil {
				t.Fatalf("expected response %v\ngot response: %v", tc.respBody, resp)
			}
		})
	}
}

func checkResponseIntegrations(response *handlers.ReturnVals, expectedResponse handlers.ReturnVals) (*handlers.ReturnVals, error) {
	if !reflect.DeepEqual(expectedResponse, response) {
		err := errors.New("DeepEqual check on responses failed")
		return response, err
	}

	return response, nil
}
