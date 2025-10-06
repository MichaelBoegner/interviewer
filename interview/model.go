package interview

import (
	"errors"
	"log/slog"
	"time"

	"github.com/michaelboegner/interviewer/billing"
	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/user"
)

type Interview struct {
	Id                      int       `json:"id"`
	ConversationID          int       `json:"conversation_id"`
	UserId                  int       `json:"user_id"`
	Length                  int       `json:"length"`
	NumberQuestions         int       `json:"number_questions"`
	NumberQuestionsAnswered int       `json:"number_questions_answered"`
	ScoreNumerator          int       `json:"score_numerator"`
	Score                   int       `json:"score"`
	Difficulty              string    `json:"difficulty"`
	Status                  string    `json:"status"`
	Language                string    `json:"language"`
	Prompt                  string    `json:"prompt"`
	JDSummary               string    `json:"jd_summary"`
	FirstQuestion           string    `json:"first_question"`
	Subtopic                string    `json:"subtopic"`
	CreatedAt               time.Time `json:"created_at,omitempty"`
	UpdatedAt               time.Time `json:"updated_at,omitempty"`
}

type Summary struct {
	ID        int       `json:"id"`
	StartedAt time.Time `json:"started_at"`
	Score     *int      `json:"score,omitempty"`
}

type InterviewService struct {
	InterviewRepo InterviewRepo
	UserRepo      user.UserRepo
	BillingRepo   billing.BillingRepo
	AI            chatgpt.AIClient
	Logger        *slog.Logger
}

var ErrNoValidCredits = errors.New("no valid credits")

func NewInterview(interviewRepo InterviewRepo, userRepo user.UserRepo, billingRepo billing.BillingRepo, ai chatgpt.AIClient, logger *slog.Logger) *InterviewService {
	return &InterviewService{
		InterviewRepo: interviewRepo,
		UserRepo:      userRepo,
		BillingRepo:   billingRepo,
		AI:            ai,
		Logger:        logger,
	}
}

type InterviewRepo interface {
	LinkConversation(interviewID, conversationID int) error
	CreateInterview(interview *Interview) (int, error)
	GetInterview(interviewID int) (*Interview, error)
	GetInterviewSummariesByUserID(userID int) ([]Summary, error)
	UpdateScore(interviewID, pointsEarned int) error
	UpdateStatus(interviewID, userID int, status string) error
}
