package mocks

import (
	"encoding/json"
	"log"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/conversation"
)

var (
	responseConversationMockCreated string
	CreatedConversationMock         *conversation.Conversation
	MessagesCreatedConversation     []conversation.Message
	now                             = time.Now().UTC()
	responseInterview               = &chatgpt.ChatGPTResponse{
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
)

func init() {
	responseConversationMarshal, err := json.Marshal(responseConversationCreated)
	if err != nil {
		log.Fatalf("MarshalResponses failed: %v", err)
	}

	responseConversationMockCreated = string(responseConversationMarshal)

	MessagesCreatedConversation = []conversation.Message{
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
		{
			ConversationID: 1,
			TopicID:        1,
			QuestionNumber: 1,
			Author:         "interviewer",
			Content:        responseConversationMockCreated,
			CreatedAt:      now,
		},
	}

}

type MockOpenAIClient struct{}

func (m *MockOpenAIClient) GetChatGPTResponseInterview(prompt string) (*chatgpt.ChatGPTResponse, error) {
	return responseInterview, nil
}

func (m *MockOpenAIClient) GetChatGPTResponseConversation(conversationHistory []map[string]string) (*chatgpt.ChatGPTResponse, error) {
	return responseConversationCreated, nil
}
