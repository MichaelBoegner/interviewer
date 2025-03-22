package interview

import (
	"time"

	"github.com/michaelboegner/interviewer/models"
)

type Interview struct {
	Id              int
	UserId          int
	Length          int
	NumberQuestions int
	Difficulty      string
	Status          string
	Score           int
	Language        string
	Prompt          string
	ChatGPTResponse *models.ChatGPTResponse
	FirstQuestion   string
	Subtopic        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OpenAIClient struct{}

type InterviewRepo interface {
	CreateInterview(interview *Interview) (int, error)
	GetInterview(interviewID int) (*Interview, error)
}

type AIClient interface {
	GetChatGPTResponse(prompt string) (*models.ChatGPTResponse, error)
}
