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
	Prompt          string
	QuestionContext *QuestionContext
	FirstQuestion   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type QuestionContext struct {
	Topic             string    `json:"topic"`
	Subtopic          string    `json:"subtopic"`
	Question          string    `json:"question"`
	Score             int       `json:"score"`
	Feedback          string    `json:"feedback"`
	NextQuestion      string    `json:"next_question"`
	MoveToNewSubtopic bool      `json:"move_to_new_subtopic"`
	MoveToNewTopic    bool      `json:"move_to_new_topic"`
	CreatedAt         time.Time `json:"created_at"`
}

type InterviewRepo interface {
	CreateInterview(interview *Interview) (int, error)
	GetInterview(interviewID int) (*Interview, error)
}
