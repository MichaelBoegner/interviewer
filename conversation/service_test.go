package conversation_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/internal/mocks"
	"github.com/michaelboegner/interviewer/interview"
)

func TestCreateConversation(t *testing.T) {
	mockAIService := &mocks.MockAIService{}

	tests := []struct {
		name           string
		interviewID    int
		conversationID int
		convo          *conversation.Conversation
		prompt         string
		firstQuestion  string
		subtopic       string
		message        string
		failRepo       bool
		expectError    bool
		expected       *conversation.Conversation
		setup          func()
	}{
		{
			name:           "CreateConversation_Success",
			interviewID:    1,
			conversationID: 1,
			convo: &conversation.Conversation{
				ID:                    1,
				InterviewID:           1,
				Topics:                conversation.ClonePredefinedTopics(),
				CurrentTopic:          1,
				CurrentSubtopic:       "Subtopic2",
				CurrentQuestionNumber: 2,
			},
			prompt:        "Prompt goes here",
			firstQuestion: "What is a goroutine?",
			subtopic:      "Concurrency",
			message:       "It's a lightweight thread",
			expectError:   false,
			expected: &conversation.Conversation{
				ID:                    1,
				InterviewID:           1,
				CurrentTopic:          1,
				Topics:                conversation.ClonePredefinedTopics(),
				CurrentSubtopic:       "Subtopic2",
				CurrentQuestionNumber: 3,
			},
			setup: func() {
				mockAIService.Scenario = mocks.ScenarioCreated
			},
		},
		{
			name:           "CreateConversation_RepoError",
			interviewID:    1,
			conversationID: 1,
			convo: &conversation.Conversation{
				ID:                    1,
				InterviewID:           1,
				CurrentTopic:          1,
				Topics:                conversation.ClonePredefinedTopics(),
				CurrentSubtopic:       "Subtopic2",
				CurrentQuestionNumber: 2,
			},
			prompt:        "Prompt",
			firstQuestion: "Question",
			subtopic:      "Subtopic",
			message:       "Answer",
			failRepo:      true,
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}
			var buf strings.Builder
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
			conversationRepo := conversation.NewMockRepo()
			interviewRepo := interview.NewMockRepo()
			conversationService := conversation.NewConvesationService(conversationRepo, interviewRepo, mockAIService, logger)
			conversationRepo.FailRepo = tc.failRepo

			convo, err := conversationService.CreateConversation(
				tc.convo,
				tc.interviewID,
				tc.prompt,
				tc.firstQuestion,
				tc.subtopic,
				tc.message)

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
	mockAIService := &mocks.MockAIService{}
	tests := []struct {
		name        string
		message     string
		interviewID int
		convo       *conversation.Conversation
		userID      int
		prompt      string
		failRepo    bool
		expectError bool
		setup       func()
	}{
		{
			name:        "AppendConversation_Success",
			message:     "T1Q2A2",
			interviewID: 1,
			convo: &conversation.Conversation{
				ID:                    1,
				InterviewID:           1,
				Topics:                conversation.ClonePredefinedTopics(),
				CurrentTopic:          1,
				CurrentSubtopic:       "Subtopic2",
				CurrentQuestionNumber: 2,
			},
			userID:      1,
			prompt:      "Prompt",
			failRepo:    false,
			expectError: false,
			setup: func() {
				mockAIService.Scenario = mocks.ScenarioAppended1
			},
		},
		{
			name:        "AppendConversation_RepoError",
			message:     "T1Q2A2",
			interviewID: 1,
			convo: &conversation.Conversation{
				ID:                    1,
				InterviewID:           1,
				Topics:                conversation.ClonePredefinedTopics(),
				CurrentTopic:          1,
				CurrentSubtopic:       "Subtopic2",
				CurrentQuestionNumber: 2,
			},
			userID:      1,
			prompt:      "Prompt",
			failRepo:    true,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}
			var buf strings.Builder
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
			conversationRepo := conversation.NewMockRepo()
			interviewRepo := interview.NewMockRepo()
			conversationService := conversation.NewConvesationService(conversationRepo, interviewRepo, mockAIService, logger)
			conversationRepo.FailRepo = tc.failRepo

			convo, err := conversationService.CreateConversation(
				tc.convo,
				tc.interviewID,
				"Prompt",
				"Question1",
				"Subtopic1",
				"T1Q1A1")

			if err != nil {
				if tc.failRepo {
					return
				}
				t.Fatalf("failed to create initial conversation: %v", err)
			}

			updatedConvo, err := conversationService.AppendConversation(
				tc.interviewID,
				tc.userID,
				convo,
				tc.message,
				tc.prompt)

			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}
			if !tc.expectError {
				if updatedConvo.CurrentTopic == 1 {
					t.Errorf("expected topic number to advance")
				}
			}
		})
	}
}
