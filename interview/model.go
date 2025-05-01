package interview

import (
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
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
	ChatGPTResponse *chatgpt.ChatGPTResponse
	FirstQuestion   string
	Subtopic        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type InterviewRepo interface {
	CreateInterview(interview *Interview) (int, error)
	GetInterview(interviewID int) (*Interview, error)
	GetInterviewsThisCycle(userID int, cycleStart, cycleEnd time.Time) (int, error)
}
