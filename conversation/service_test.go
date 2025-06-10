package conversation_test

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/internal/mocks"
	"github.com/michaelboegner/interviewer/interview"
)

func TestCreateConversation(t *testing.T) {
	tests := []struct {
		name           string
		interviewID    int
		conversationID int
		prompt         string
		firstQuestion  string
		subtopic       string
		message        string
		failRepo       bool
		expectError    bool
		expected       *conversation.Conversation
	}{
		{
			name:           "CreateConversation_Success",
			interviewID:    1,
			conversationID: 1,
			prompt:         "Prompt goes here",
			firstQuestion:  "What is a goroutine?",
			subtopic:       "Concurrency",
			message:        "It's a lightweight thread",
			expectError:    false,
			expected: &conversation.Conversation{
				InterviewID:           1,
				CurrentTopic:          1,
				CurrentSubtopic:       "Subtopic2",
				CurrentQuestionNumber: 2,
			},
		},
		{
			name:           "CreateConversation_RepoError",
			interviewID:    1,
			conversationID: 1,
			prompt:         "Prompt",
			firstQuestion:  "Question",
			subtopic:       "Subtopic",
			message:        "Answer",
			failRepo:       true,
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			repo := conversation.NewMockRepo()
			interviewRepo := interview.NewMockRepo()
			if tc.failRepo {
				repo.FailRepo = true
			}

			ai := &mocks.MockOpenAIClient{}

			convo, err := conversation.CreateConversation(repo, interviewRepo, ai, tc.interviewID, tc.conversationID, tc.prompt, tc.firstQuestion, tc.subtopic, tc.message)

			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}
			if !tc.expectError {
				expected := tc.expected
				got := convo

				if diff := cmp.Diff(expected, got,
					cmpopts.IgnoreFields(conversation.Conversation{}, "ID", "Topics", "CreatedAt", "UpdatedAt"),
				); diff != "" {
					t.Errorf("Conversation mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestAppendConversation(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		interviewID    int
		conversationID int
		userID         int
		prompt         string
		failRepo       bool
		expectError    bool
	}{
		{
			name:           "AppendConversation_Success",
			message:        "Answer1",
			interviewID:    1,
			conversationID: 1,
			userID:         1,
			prompt:         "Prompt",
			failRepo:       false,
			expectError:    false,
		},
		{
			name:           "AppendConversation_RepoError",
			message:        "Answer1",
			interviewID:    1,
			conversationID: 1,
			userID:         1,
			prompt:         "Prompt",
			failRepo:       true,
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			log.SetOutput(&buf)
			defer showLogsIfFail(t, tc.name, buf)

			repo := conversation.NewMockRepo()
			interviewRepo := interview.NewMockRepo()
			if tc.failRepo {
				repo.FailRepo = true
			}
			ai := &mocks.MockOpenAIClient{}

			convo, err := conversation.CreateConversation(repo, interviewRepo, ai, tc.interviewID, tc.conversationID, "Prompt", "Question1", "Subtopic1", "Answer1")
			if err != nil {
				if tc.failRepo {
					return
				}
				t.Fatalf("failed to create initial conversation: %v", err)
			}

			updatedConvo, err := conversation.AppendConversation(repo, interviewRepo, ai, tc.interviewID, tc.userID, convo, tc.message, tc.prompt)

			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}
			if !tc.expectError {
				if updatedConvo.CurrentTopic == 1 {
					t.Errorf("expected question number to advance but got 1")
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
