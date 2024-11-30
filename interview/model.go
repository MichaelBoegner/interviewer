package interview

import (
	"time"
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
	QuestionContext *QuestionContext
	FirstQuestion   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type QuestionContext struct {
	Topic             string
	Subtopic          string
	Question          string
	Score             int
	Feedback          string
	NextQuestion      string
	MoveToNewSubtopic bool
	MoveToNewTopic    bool
}

type InterviewRepo interface {
	CreateInterview(interview *Interview) (int, error)
	GetInterview(interviewID int) (*Interview, error)
}
