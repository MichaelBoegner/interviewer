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
	QuestionContext *models.QuestionContext
	FirstQuestion   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type InterviewRepo interface {
	CreateInterview(interview *Interview) (int, error)
	GetInterview(interviewID int) (*Interview, error)
}
