package conversation

import "time"

type Author string

const (
	AuthorInterviewer Author = "interviewer"
	User              Author = "user"
)

const (
	TopicIntroduction = 1
	TopicCoding       = 2
	TopicSystemDesign = 3
	TopicDatabases    = 4
	TopicBehavioral   = 5
	TopicBackend      = 6
)

var PredefinedTopics = map[int]Topic{
	TopicIntroduction: {ID: TopicIntroduction, Name: "Introduction"},
	TopicCoding:       {ID: TopicCoding, Name: "Coding"},
	TopicSystemDesign: {ID: TopicSystemDesign, Name: "System Design"},
	TopicDatabases:    {ID: TopicDatabases, Name: "Databases and Data Management"},
	TopicBehavioral:   {ID: TopicBehavioral, Name: "Behavioral"},
	TopicBackend:      {ID: TopicBackend, Name: "General Backend Knowledge"},
}

type Conversation struct {
	ID          int           `json:"id"`
	InterviewID int           `json:"interview_id"`
	Topics      map[int]Topic `json:"topics"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type Topic struct {
	ID             int              `json:"id"`
	ConversationID int              `json:"conversation_id"`
	Name           string           `json:"name"`
	Questions      map[int]Question `json:"questions"`
}

type Question struct {
	ID             int       `json:"id"`
	QuestionNumber int       `json:"question_number"`
	ConversationID int       `json:"conversation_id"`
	TopicID        int       `json:"topic_id"`
	Prompt         string    `json:"prompt"`
	Messages       []Message `json:"messages"`
	CreatedAt      time.Time `json:"created_at"`
}

type Message struct {
	ID         int       `json:"id"`
	QuestionID int       `json:"question_id"`
	Author     Author    `json:"author"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type ConversationRepo interface {
	CheckForConversation(interviewID int) bool
	GetConversation(interviewID int) (*Conversation, error)
	CreateConversation(conversation *Conversation) (int, error)
	CreateQuestion(conversation *Conversation, prompt string) (int, error)
	GetQuestion(Conversation *Conversation) (*Question, error)
	CreateMessages(conversation *Conversation, messages []Message) error
	AddMessage(questionID int, message *Message) (int, error)
	GetMessages(questionID int) ([]Message, error)
}
