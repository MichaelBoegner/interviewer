package conversation

import "time"

type Author string

const (
	AuthorInterviewer Author = "Interviewer"
	User              Author = "User"
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
	ID             int            `json:"id"`
	ConversationID int            `json:"conversation_id"`
	Name           string         `json:"name"`
	Questions      map[int]string `json:"questions"`
}

type Question struct {
	QuestionNumber int       `json:"question_number"`
	Messages       []Message `json:"messages"`
	CreatedAt      time.Time `json:"created_at"`
}

type Message struct {
	Author    Author    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ConversationRepo interface {
	CreateConversation(conversation *Conversation) (int, error)
}
