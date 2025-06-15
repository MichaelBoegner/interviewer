package chatgpt

import (
	"fmt"
	"os"
)

type ChatGPTResponse struct {
	Topic        string `json:"topic"`
	Subtopic     string `json:"subtopic"`
	Question     string `json:"question"`
	Score        int    `json:"score"`
	Feedback     string `json:"feedback"`
	NextQuestion string `json:"next_question"`
	NextTopic    string `json:"next_topic"`
	NextSubtopic string `json:"next_subtopic"`
}

type OpenAIClient struct {
	APIKey string
}

type OpenAIError struct {
	StatusCode int
	Message    string
}

func (e *OpenAIError) Error() string {
	return fmt.Sprintf("OpenAI error %d: %s", e.StatusCode, e.Message)
}

func NewOpenAI() *OpenAIClient {
	return &OpenAIClient{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}
}

type AIClient interface {
	GetChatGPTResponseInterview(prompt string) (*ChatGPTResponse, error)
	GetChatGPTResponseConversation(conversationHistory []map[string]string) (*ChatGPTResponse, error)
}
