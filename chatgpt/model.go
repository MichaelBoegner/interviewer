package chatgpt

import (
	"os"
	"time"
)

type ChatGPTResponse struct {
	Topic        string    `json:"topic"`
	Subtopic     string    `json:"subtopic"`
	Question     string    `json:"question"`
	Score        int       `json:"score"`
	Feedback     string    `json:"feedback"`
	NextQuestion string    `json:"next_question"`
	NextTopic    string    `json:"next_topic"`
	NextSubtopic string    `json:"next_subtopic"`
	CreatedAt    time.Time `json:"created_at"`
}

type OpenAIClient struct {
	APIKey string
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
