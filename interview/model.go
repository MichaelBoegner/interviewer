package interview

import (
	"errors"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
)

type Interview struct {
	Id                      int
	UserId                  int
	Length                  int
	NumberQuestions         int
	NumberQuestionsAnswered int
	ScoreNumerator          int
	Score                   int
	Difficulty              string
	Status                  string
	Language                string
	Prompt                  string
	ChatGPTResponse         *chatgpt.ChatGPTResponse
	FirstQuestion           string
	Subtopic                string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type Summary struct {
	ID        int       `json:"id"`
	StartedAt time.Time `json:"started_at"`
	Score     *int      `json:"score,omitempty"`
}

var ErrNoValidCredits = errors.New("no valid credits")

type InterviewRepo interface {
	CreateInterview(interview *Interview) (int, error)
	GetInterview(interviewID int) (*Interview, error)
	GetInterviewSummariesByUserID(userID int) ([]Summary, error)
	UpdateScore(interviewID, pointsEarned int) error
}
