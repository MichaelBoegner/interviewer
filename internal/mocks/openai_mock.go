package mocks

import (
	"encoding/json"
	"log"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/conversation"
)

var (
	responseConversationMockCreated   string
	responseConversationMockAppended  string
	CreatedConversationMock           *conversation.Conversation
	MessagesCreatedConversationT1Q1   []conversation.Message
	MessagesCreatedConversationT1Q2   []conversation.Message
	MessagesCreatedConversationT1Q2A2 []conversation.Message
	MessagesAppendedConversationT2Q1  []conversation.Message
	now                               = time.Now().UTC()
	responseInterview                 = &chatgpt.ChatGPTResponse{
		Topic:        "None",
		Subtopic:     "None",
		Question:     "None",
		Score:        0,
		Feedback:     "None",
		NextQuestion: "Question1",
		NextTopic:    "Introduction",
		NextSubtopic: "Subtopic1",
		CreatedAt:    now,
	}
	responseConversationCreated = &chatgpt.ChatGPTResponse{
		Topic:        "Introduction",
		Subtopic:     "Subtopic1",
		Question:     "Question1",
		Score:        10,
		Feedback:     "Feedback1",
		NextQuestion: "Question2",
		NextTopic:    "Introduction",
		NextSubtopic: "Subtopic2",
		CreatedAt:    now,
	}
	responseConversationAppended = &chatgpt.ChatGPTResponse{
		Topic:        "Introduction",
		Subtopic:     "Subtopic2",
		Question:     "Question2",
		Score:        10,
		Feedback:     "Feedback2",
		NextQuestion: "Question1",
		NextTopic:    "Coding",
		NextSubtopic: "Subtopic1",
		CreatedAt:    now,
	}
)

func init() {
	responseConversationCreatedMarshal, err := json.Marshal(responseConversationCreated)
	if err != nil {
		log.Fatalf("MarshalResponses failed: %v", err)
	}

	responseConversationMockCreated = string(responseConversationCreatedMarshal)

	responseConversationAppendedMarshal, err := json.Marshal(responseConversationAppended)
	if err != nil {
		log.Fatalf("MarshalResponses failed: %v", err)
	}

	responseConversationMockAppended = string(responseConversationAppendedMarshal)

	MessagesCreatedConversationT1Q1 = []conversation.Message{
		{
			ConversationID: 1,
			TopicID:        1,
			QuestionNumber: 1,
			Author:         "system",
			Content:        TestPrompt,
			CreatedAt:      now,
		},
		{
			ConversationID: 1,
			TopicID:        1,
			QuestionNumber: 1,
			Author:         "interviewer",
			Content:        "Question1",
			CreatedAt:      now,
		},
		{
			ConversationID: 1,
			TopicID:        1,
			QuestionNumber: 1,
			Author:         "user",
			Content:        "Answer1",
			CreatedAt:      now,
		},
	}

	MessagesCreatedConversationT1Q2 = []conversation.Message{
		{
			ConversationID: 1,
			TopicID:        1,
			QuestionNumber: 2,
			Author:         "interviewer",
			Content:        responseConversationMockCreated,
			CreatedAt:      now,
		},
	}

	MessagesCreatedConversationT1Q2A2 = []conversation.Message{
		{
			ConversationID: 1,
			TopicID:        1,
			QuestionNumber: 2,
			Author:         "user",
			Content:        "Answer2",
			CreatedAt:      now,
		},
	}

	MessagesAppendedConversationT2Q1 = []conversation.Message{
		{
			ConversationID: 1,
			TopicID:        2,
			QuestionNumber: 1,
			Author:         "system",
			Content:        TestPrompt,
			CreatedAt:      now,
		},
		{
			ConversationID: 1,
			TopicID:        2,
			QuestionNumber: 1,
			Author:         "interviewer",
			Content:        responseConversationMockAppended,
			CreatedAt:      now,
		},
	}

}

type MockOpenAIClient struct{}

func (m *MockOpenAIClient) GetChatGPTResponseInterview(prompt string) (*chatgpt.ChatGPTResponse, error) {
	return responseInterview, nil
}

func (m *MockOpenAIClient) GetChatGPTResponseConversation(conversationHistory []map[string]string) (*chatgpt.ChatGPTResponse, error) {
	if len(conversationHistory) == 3 {
		return responseConversationCreated, nil
	}

	return responseConversationAppended, nil
}
