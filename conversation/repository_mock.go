package conversation

import "time"

type MockRepo struct{}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (repo *MockRepo) CheckForConversation(interviewID int) bool {
	return false
}

func (repo *MockRepo) GetConversation(interviewID int) (*Conversation, error) {
	conversationResponse := &Conversation{
		ID:          1,
		InterviewID: 1,
		Topics:      PredefinedTopics,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	topic := conversationResponse.Topics[1]
	topic.ConversationID = 1
	topic.Questions = make(map[int]*Question)

	question := topic.Questions[1]
	question.ID = 1
	question.QuestionNumber = 1
	question.Prompt = "What is the flight speed of an unladdened swallow?"

	messageFirst := &Message{
		ID:         1,
		QuestionID: 1,
		Author:     "interviewer",
		Content:    "What is the flight speed of an unladdened swallow?",
		CreatedAt:  time.Now(),
	}

	messageResponse := &Message{
		ID:         2,
		QuestionID: 1,
		Author:     "user",
		Content:    "European or African?",
		CreatedAt:  time.Now(),
	}

	messageResponse.ID = 2
	messageResponse.QuestionID = 1
	messageResponse.CreatedAt = time.Now()

	question.Messages = make([]Message, 0)
	question.Messages = append(question.Messages, *messageFirst)
	question.Messages = append(question.Messages, *messageResponse)

	conversationResponse.Topics[1] = topic
	conversationResponse.Topics[1].Questions[1] = question

	return conversationResponse, nil
}

func (repo *MockRepo) CreateConversation(conversation *Conversation) (int, error) {
	return 1, nil
}

func (repo *MockRepo) CreateQuestion(conversation *Conversation) (int, error) {
	return 1, nil
}

func (repo *MockRepo) CreateMessages(conversation *Conversation, messages []Message) error {
	return nil
}

func (repo *MockRepo) AddMessage(questionID int, message *Message) (int, error) {
	return 3, nil
}
