package mocks

import (
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
)

type MockOpenAIClient struct{}

func (m *MockOpenAIClient) GetChatGPTResponseInterview(prompt string) (*chatgpt.ChatGPTResponse, error) {
	return &chatgpt.ChatGPTResponse{
		Topic:        "None",
		Subtopic:     "None",
		Question:     "None",
		Score:        0,
		Feedback:     "None",
		NextQuestion: "Tell me a little bit about your work history.",
		NextTopic:    "Introduction",
		NextSubtopic: "General Background",
		CreatedAt:    time.Now(),
	}, nil
}

func (m *MockOpenAIClient) GetChatGPTResponseConversation(conversationHistory []map[string]string) (*chatgpt.ChatGPTResponse, error) {
	return &chatgpt.ChatGPTResponse{
		Topic:        "Introduction",
		Subtopic:     "General Background",
		Question:     "Tell me a little bit about your work history",
		Score:        10,
		Feedback:     "Sounds like you have a good deal of problem solving experience.",
		NextQuestion: "Can you tell me about your most recent backend project?",
		NextTopic:    "Introduction",
		NextSubtopic: "General Engineering Experience",
		CreatedAt:    time.Now(),
	}, nil
}
