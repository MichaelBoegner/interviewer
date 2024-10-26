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
	ID          int
	InterviewID int
	Topics      map[int]Topic
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Topic struct {
	ID             int
	ConversationID int
	Name           string
	Questions      map[int]string
}

type Question struct {
	QuestionNumber int
	Messages       []Message
	CreatedAt      time.Time
}

type Message struct {
	Author    Author
	Content   string
	CreatedAt time.Time
}

type ConversationRepo interface {
	CreateConversation(conversation *Conversation) (int, error)
}
