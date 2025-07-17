package embedding

import (
	"context"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
)

type EmbedInput struct {
	InterviewID    int
	ConversationID int
	MessageID      int
	TopicID        int
	QuestionNumber int
	Question       string
	UserResponse   string
	CreatedAt      time.Time
}

type Embedding struct {
	ID             int
	InterviewID    int
	ConversationID int
	MessageID      int
	TopicID        int
	QuestionNumber int
	Summary        string
	Vector         []float32
	CreatedAt      time.Time
}

type Service struct {
	Repo       Repository
	Embedder   Embedder
	Summarizer Summarizer
}

type Repository interface {
	StoreEmbedding(ctx context.Context, e Embedding) error
	GetSimilarEmbeddings(ctx context.Context, interviewID, topicID, questionNumber, excludeMessageID int, queryVec []float32, limit int) ([]string, error)
}

type Embedder interface {
	EmbedText(ctx context.Context, input string) ([]float32, error)
}

type Summarizer interface {
	ExtractResponseSummary(question, response string) (*chatgpt.ChatGPTResponse, error)
}

func NewService(repo Repository, embedder Embedder, summarizer Summarizer) *Service {
	return &Service{
		Repo:       repo,
		Embedder:   embedder,
		Summarizer: summarizer,
	}
}
