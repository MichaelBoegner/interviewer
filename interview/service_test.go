package interview

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/michaelboegner/interviewer/internal/mocks"
)

func TestStartInterview(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		length       int
		numQuestions int
		difficulty   string
		aiClient     *mocks.MockOpenAIClient
		failRepo     bool
		expected     *Interview
		expectError  bool
	}{
		{
			name:         "StartInterview_Success",
			userID:       1,
			length:       30,
			numQuestions: 3,
			difficulty:   "easy",
			aiClient:     &mocks.MockOpenAIClient{},
			expected: &Interview{
				UserId:          1,
				Length:          30,
				NumberQuestions: 3,
				Difficulty:      "easy",
				Status:          "Running",
				Score:           100,
				Language:        "Python",
				FirstQuestion:   "Question1",
				Subtopic:        "None",
			},
			expectError: false,
		},
		{
			name:         "StartInterview_RepoError",
			userID:       1,
			length:       30,
			numQuestions: 3,
			difficulty:   "easy",
			aiClient:     &mocks.MockOpenAIClient{},
			failRepo:     true,
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			repo := NewMockRepo()
			if tc.failRepo {
				repo.failRepo = true
			}

			interview, err := StartInterview(repo, tc.aiClient, tc.userID, tc.length, tc.numQuestions, tc.difficulty)

			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}

			if !tc.expectError {
				expected := tc.expected
				got := interview

				if diff := cmp.Diff(expected, got,
					cmpopts.IgnoreFields(Interview{}, "Id", "CreatedAt", "UpdatedAt", "Prompt", "ChatGPTResponse"),
				); diff != "" {
					t.Errorf("Interview mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestGetInterview(t *testing.T) {
	tests := []struct {
		name        string
		interviewID int
		setup       *Interview
		failRepo    bool
		expected    *Interview
		expectError bool
	}{
		{
			name:        "GetInterview_Success",
			interviewID: 1,
			setup: &Interview{
				Id:              1,
				UserId:          1,
				Length:          30,
				NumberQuestions: 3,
				Difficulty:      "easy",
				Status:          "Running",
				Score:           100,
				Language:        "Python",
				FirstQuestion:   "Question1",
				Subtopic:        "None",
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			},
			expected: &Interview{
				Id:              1,
				UserId:          1,
				Length:          30,
				NumberQuestions: 2,
				Difficulty:      "easy",
				Status:          "running",
				Score:           0,
				Language:        "python",
				FirstQuestion:   "question1",
				Subtopic:        "None",
			},
			expectError: false,
		},
		{
			name:        "GetInterview_RepoError",
			interviewID: 1,
			failRepo:    true,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			repo := NewMockRepo()
			if tc.failRepo {
				repo.failRepo = true
			}
			if tc.setup != nil {
				_, err := repo.CreateInterview(tc.setup)
				if err != nil {
					t.Logf("CreateInterview failed: %v", err)
				}
			}

			got, err := GetInterview(repo, tc.interviewID)

			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}

			if !tc.expectError {
				expected := tc.expected

				if diff := cmp.Diff(expected, got,
					cmpopts.IgnoreFields(Interview{}, "CreatedAt", "UpdatedAt"),
				); diff != "" {
					t.Errorf("Interview mismatch (-want +got):\n%s", diff)
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
