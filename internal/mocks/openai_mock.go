package mocks

import (
	"encoding/json"
	"log"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/conversation"
)

var (
	responseConversationMockCreated  string
	responseConversationMockAppended string
	CreatedConversationMock          *conversation.Conversation
	MessagesCreatedConversationT1Q1  []conversation.Message
	MessagesCreatedConversationT1Q2  []conversation.Message
	MessagesAppendedConversationT2Q1 []conversation.Message
	now                              = time.Now().UTC()
	responseInterview                = &chatgpt.ChatGPTResponse{
		Topic:        "None",
		Subtopic:     "None",
		Question:     "None",
		Score:        0,
		Feedback:     "None",
		NextQuestion: "Tell me a little bit about your work history.",
		NextTopic:    "Introduction",
		NextSubtopic: "General Background",
		CreatedAt:    now,
	}
	responseConversationCreated = &chatgpt.ChatGPTResponse{
		Topic:        "Introduction",
		Subtopic:     "General Background",
		Question:     "Tell me a little bit about your work history",
		Score:        10,
		Feedback:     "Sounds like you have a good deal of problem solving experience.",
		NextQuestion: "Can you tell me about your most recent backend project?",
		NextTopic:    "Introduction",
		NextSubtopic: "General Engineering Experience",
		CreatedAt:    now,
	}
	responseConversationAppended = &chatgpt.ChatGPTResponse{
		Topic:        "Introduction",
		Subtopic:     "General Engineering Experience",
		Question:     "Can you tell me about your most recent backend project?",
		Score:        10,
		Feedback:     "Great job building something more than just a toy project!",
		NextQuestion: "Can you write me a func to reverse a string?",
		NextTopic:    "Coding",
		NextSubtopic: "String Alogrithms",
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
			Content:        "Tell me a little bit about your work history.",
			CreatedAt:      now,
		},
		{
			ConversationID: 1,
			TopicID:        1,
			QuestionNumber: 1,
			Author:         "user",
			Content:        "I have been a TSE for 5 years.",
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
		{
			ConversationID: 1,
			TopicID:        1,
			QuestionNumber: 2,
			Author:         "user",
			Content:        "I built a mock interviewer app in Go.",
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
