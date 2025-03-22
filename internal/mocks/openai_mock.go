package mocks

import (
	"time"

	"github.com/michaelboegner/interviewer/models"
)

type MockOpenAIClient struct{}

func (m *MockOpenAIClient) GetChatGPTResponse(prompt string) (*models.ChatGPTResponse, error) {
	return &models.ChatGPTResponse{
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
