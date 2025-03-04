package conversation

import "time"

type Author string

const (
	System      Author = "system"
	Interviewer Author = "interviewer"
	User        Author = "user"
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
	ID                    int           `json:"id"`
	InterviewID           int           `json:"interview_id"`
	Topics                map[int]Topic `json:"topics"`
	CurrentTopic          int           `json:"current_topic"`
	CurrentSubtopic       string        `json:"current_subtopic"`
	CurrentQuestionNumber int           `json: "current_question_number"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
}

type Topic struct {
	ID             int               `json:"id"`
	ConversationID int               `json:"conversation_id"`
	Name           string            `json:"name"`
	Questions      map[int]*Question `json:"questions"`
}

type Question struct {
	ConversationID int       `json:"conversation_id"`
	TopicID        int       `json:"topic_id"`
	QuestionNumber int       `json:"question_number"`
	Prompt         string    `json:"prompt"`
	Messages       []Message `json:"messages"`
	CreatedAt      time.Time `json:"created_at"`
}

type Message struct {
	ConversationID int       `json:"conversation_id"`
	TopicID        int       `json:"topic_id"`
	QuestionNumber int       `json:"question_id"`
	Author         Author    `json:"author"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

type ConversationRepo interface {
	CheckForConversation(interviewID int) bool
	GetConversation(interviewID int) (*Conversation, error)
	CreateConversation(conversation *Conversation) (int, error)
	UpdateConversationCurrents(conversationID, currentQuestionNumber, topicID int, subtopic string) (int, error)
	CreateQuestion(conversation *Conversation, prompt string) (int, error)
	AddQuestion(question *Question) (int, error)
	GetQuestions(Conversation *Conversation) ([]*Question, error)
	CreateMessages(conversation *Conversation, messages []Message) error
	AddMessage(conversationID, topic_id, questionNumber int, message *Message) (int, error)
	GetMessages(conversationID, topic_id, questionNumber int) ([]Message, error)
}
