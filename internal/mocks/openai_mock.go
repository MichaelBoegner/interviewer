package mocks

import (
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
)

type MockOpenAIClient struct{}

func (m *MockOpenAIClient) GetChatGPTResponseInterview(prompt string) (*chatgpt.ChatGPTResponse, error) {
	return &chatgpt.ChatGPTResponse{
		Topic:        "Introduction",
		Subtopic:     "General Background",
		Question:     "None",
		Score:        0,
		Feedback:     "None",
		NextQuestion: "Tell me a little bit about your work history.",
		NextTopic:    "Introduction",
		NextSubtopic: "General Backend Experience",
		CreatedAt:    time.Now(),
	}, nil
}

func (m *MockOpenAIClient) GetChatGPTResponseConversation(conversationHistory []map[string]string) (*chatgpt.ChatGPTResponse, error) {
	return &chatgpt.ChatGPTResponse{
		Topic:        "Introduction",
		Subtopic:     "General Background",
		Question:     "None",
		Score:        0,
		Feedback:     "None",
		NextQuestion: "Tell me a little bit about your work history.",
		NextTopic:    "Introduction",
		NextSubtopic: "General Backend Experience",
		CreatedAt:    time.Now(),
	}, nil
}
