package interview

import (
	"errors"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
)

type Interview struct {
	Id                      int                      `json:"id"`
	ConversationID          int                      `json:"conversation_id"`
	UserId                  int                      `json:"user_id"`
	Length                  int                      `json:"length"`
	NumberQuestions         int                      `json:"number_questions"`
	NumberQuestionsAnswered int                      `json:"number_questions_answered"`
	ScoreNumerator          int                      `json:"score_numerator"`
	Score                   int                      `json:"score"`
	Difficulty              string                   `json:"difficulty"`
	Status                  string                   `json:"status"`
	Language                string                   `json:"language"`
	Prompt                  string                   `json:"prompt"`
	ChatGPTResponse         *chatgpt.ChatGPTResponse `json:"chatgpt_response,omitempty"`
	FirstQuestion           string                   `json:"first_question"`
	Subtopic                string                   `json:"subtopic"`
	CreatedAt               time.Time                `json:"created_at,omitempty"`
	UpdatedAt               time.Time                `json:"updated_at,omitempty"`
}

type Summary struct {
	ID        int       `json:"id"`
	StartedAt time.Time `json:"started_at"`
	Score     *int      `json:"score,omitempty"`
}

var ErrNoValidCredits = errors.New("no valid credits")

type InterviewRepo interface {
	LinkConversation(interviewID, conversationID int) error
	CreateInterview(interview *Interview) (int, error)
	GetInterview(interviewID int) (*Interview, error)
	GetInterviewSummariesByUserID(userID int) ([]Summary, error)
	UpdateScore(interviewID, pointsEarned int) error
	UpdateStatus(interviewID, userID int, status string) error
	UpdateCreatedInterview(interview *Interview) error
}
