package handlers

import (
	"database/sql"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

type Handler struct {
	UserRepo         user.UserRepo
	InterviewRepo    interview.InterviewRepo
	ConversationRepo conversation.ConversationRepo
	TokenRepo        token.TokenRepo
	OpenAI           chatgpt.AIClient
	DB               *sql.DB
}

type ReturnVals struct {
	ID             int                        `json:"id,omitempty"`
	UserID         int                        `json:"user_id,omitempty"`
	InterviewID    int                        `json:"interview_id,omitempty"`
	ConversationID int                        `json:"conversation_id,omitempty"`
	Body           string                     `json:"body,omitempty"`
	Username       string                     `json:"username,omitempty"`
	Email          string                     `json:"email,omitempty"`
	FirstQuestion  string                     `json:"first_question,omitempty"`
	NextQuestion   string                     `json:"next_question,omitempty"`
	JWToken        string                     `json:"jwtoken,omitempty"`
	RefreshToken   string                     `json:"refresh_token,omitempty"`
	Error          string                     `json:"error,omitempty"`
	Users          map[int]user.User          `json:"users,omitempty"`
	Conversation   *conversation.Conversation `json:"conversation,omitempty"`
	User           *user.User                 `json:"user,omitempty"`
}

func NewHandler(
	interviewRepo interview.InterviewRepo,
	userRepo user.UserRepo,
	tokenRepo token.TokenRepo,
	conversationRepo conversation.ConversationRepo,
	openAI chatgpt.AIClient,
	db *sql.DB) *Handler {
	return &Handler{
		InterviewRepo:    interviewRepo,
		UserRepo:         userRepo,
		TokenRepo:        tokenRepo,
		ConversationRepo: conversationRepo,
		OpenAI:           openAI,
		DB:               db,
	}
}
