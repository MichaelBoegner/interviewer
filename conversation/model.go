package conversation

import "time"

type Conversation struct {
	ID          int
	InterviewID int
	Messages    map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ConversationRepo interface {
	CreateConversation(conversation *Conversation) (int, error)
}
