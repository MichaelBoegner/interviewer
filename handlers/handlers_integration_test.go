package handlers_test

import (
	"net/http"
	"testing"

	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/internal/testutil"
	"github.com/michaelboegner/interviewer/middleware"
)

func TestInterviewsHandler_Post_Integration(t *testing.T) {
	// t.Skip("TODO: skipping for now while setting up recorded demo")

	_, _, err := testutil.CreateTestUserAndJWT(t)
	if err != nil {
		t.Fatalf("CreateTestUserAndJWT failed: %v", err)
	}

	tests := []TestCase{
		{
			name:           "CreateInterview_Success",
			reqBody:        `{}`,
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
			// user := user.NewMockRepo()
			// handler := &handlers.Handler{UserRepo: mockUserRepo}

			// w, req, tc, err := setRequestAndWriter(http.MethodPost, "/api/users", tc)
			// if err != nil {
			// 	t.Fatalf("failed to set request and writer")
			// }

			// Act
			// handler.UsersHandler(w, req)

			// Assert
			// if w.Code != tc.expectedStatus {
			// 	t.Fatalf("expected status %d, got %d", tc.expectedStatus, w.Code)
			// }

			// // Validate resp
			// resp, err := checkResponse(w, tc.respBody, tc.expectError)
			// if err != nil {
			// 	t.Fatalf("expected response %v and error %v\ngot response: %v and error %v", tc.respBody, tc.expectError, resp, resp.Error)
			// }
		})
	}
}
