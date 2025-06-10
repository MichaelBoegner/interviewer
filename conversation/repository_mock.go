package conversation

import (
	"errors"
	"time"
)

type MockRepo struct {
	FailRepo bool
}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (m *MockRepo) CheckForConversation(interviewID int) (bool, error) {
	if m.FailRepo {
		return false, errors.New("Mocked DB failure")
	}

	return true, nil
}

func (m *MockRepo) GetConversation(interviewID int) (*Conversation, error) {
	if m.FailRepo {
		return nil, errors.New("Mocked DB failure")
	}

	conversationResponse := &Conversation{
		ID:          1,
		InterviewID: 1,
		Topics:      PredefinedTopics,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	topic := conversationResponse.Topics[1]
	topic.ConversationID = 1
	topic.Questions = make(map[int]*Question)

	question := topic.Questions[1]
	question.ConversationID = 1
	question.QuestionNumber = 1
	question.Prompt = "What is the flight speed of an unladdened swallow?"

	messageFirst := &Message{
		ConversationID: 1,
		QuestionNumber: 1,
		Author:         "interviewer",
		Content:        "What is the flight speed of an unladdened swallow?",
		CreatedAt:      time.Now().UTC(),
	}

	messageResponse := &Message{
		ConversationID: 1,
		QuestionNumber: 1,
		Author:         "user",
		Content:        "European or African?",
		CreatedAt:      time.Now().UTC(),
	}

	question.Messages = make([]Message, 0)
	question.Messages = append(question.Messages, *messageFirst)
	question.Messages = append(question.Messages, *messageResponse)

	conversationResponse.Topics[1] = topic
	conversationResponse.Topics[1].Questions[1] = question

	return conversationResponse, nil
}

func (m *MockRepo) CreateConversation(interviewId int, conversation *Conversation) error {
	if m.FailRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) CreateQuestion(conversation *Conversation, prompt string) (int, error) {
	if m.FailRepo {
		return 0, errors.New("Mocked DB failure")
	}

	return 1, nil
}

func (m *MockRepo) CreateMessages(conversation *Conversation, messages []Message) error {
	if m.FailRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) AddMessage(conversationID, topic_id, questionNumber int, message Message) (int, error) {
	if m.FailRepo {
		return 0, errors.New("Mocked DB failure")
	}

	return 3, nil
}

func (m *MockRepo) AddQuestion(question *Question) (int, error) {
	if m.FailRepo {
		return 0, errors.New("Mocked DB failure")
	}

	return 2, nil
}

func (m *MockRepo) GetMessages(conversationID, topic_id, questionNumber int) ([]Message, error) {
	if m.FailRepo {
		return nil, errors.New("Mocked DB failure")
	}

	var messages = []Message{}
	return messages, nil
}

func (m *MockRepo) GetQuestions(Conversation *Conversation) ([]*Question, error) {
	if m.FailRepo {
		return nil, errors.New("Mocked DB failure")
	}

	var questions = []*Question{}
	return questions, nil
}

func (m *MockRepo) UpdateConversationCurrents(conversationID, currentQuestionNumber, topicID int, subtopic string) (int, error) {
	if m.FailRepo {
		return 0, errors.New("Mocked DB failure")
	}

	return 1, nil
}
