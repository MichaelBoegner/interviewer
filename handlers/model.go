package handlers

import (
	"database/sql"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/customerio"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

type PasswordResetRequest struct {
	Email string `json:"email"`
}

type PasswordResetPayload struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
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

type Handler struct {
	UserRepo         user.UserRepo
	InterviewRepo    interview.InterviewRepo
	ConversationRepo conversation.ConversationRepo
	TokenRepo        token.TokenRepo
	Mailer           customerio.Mailer
	OpenAI           chatgpt.AIClient
	DB               *sql.DB
}

func NewHandler(
	interviewRepo interview.InterviewRepo,
	userRepo user.UserRepo,
	tokenRepo token.TokenRepo,
	conversationRepo conversation.ConversationRepo,
	mailer customerio.Mailer,
	openAI chatgpt.AIClient,
	db *sql.DB) *Handler {
	return &Handler{
		InterviewRepo:    interviewRepo,
		UserRepo:         userRepo,
		TokenRepo:        tokenRepo,
		ConversationRepo: conversationRepo,
		Mailer:           mailer,
		OpenAI:           openAI,
		DB:               db,
	}
}
