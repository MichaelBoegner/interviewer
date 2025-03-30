package mocks

import (
	"encoding/json"
	"log"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
)

var responseConversation = &chatgpt.ChatGPTResponse{
	Topic:        "Introduction",
	Subtopic:     "General Background",
	Question:     "Tell me a little bit about your work history",
	Score:        10,
	Feedback:     "Sounds like you have a good deal of problem solving experience.",
	NextQuestion: "Can you tell me about your most recent backend project?",
	NextTopic:    "Introduction",
	NextSubtopic: "General Engineering Experience",
	CreatedAt:    time.Now(),
}

var responseInterview = &chatgpt.ChatGPTResponse{
	Topic:        "None",
	Subtopic:     "None",
	Question:     "None",
	Score:        0,
	Feedback:     "None",
	NextQuestion: "Tell me a little bit about your work history.",
	NextTopic:    "Introduction",
	NextSubtopic: "General Background",
	CreatedAt:    time.Now(),
}

func MarshalResponses(response *chatgpt.ChatGPTResponse) (string, error) {
	responseConversationString, err := json.Marshal(responseConversation)
	if err != nil {
		log.Fatalf("MarshalResponses failed: %v", err)
		return "", err
	}
	return string(responseConversationString), nil
}

type MockOpenAIClient struct{}

func (m *MockOpenAIClient) GetChatGPTResponseInterview(prompt string) (*chatgpt.ChatGPTResponse, error) {
	return responseInterview, nil
}

func (m *MockOpenAIClient) GetChatGPTResponseConversation(conversationHistory []map[string]string) (*chatgpt.ChatGPTResponse, error) {
	return responseConversation, nil
}
