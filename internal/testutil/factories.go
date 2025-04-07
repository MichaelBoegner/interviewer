package testutil

import (
	"time"

	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/internal/mocks"
)

type ConversationBuilder struct {
	convo *conversation.Conversation
}

func NewConversationBuilder() *ConversationBuilder {
	now := time.Now().UTC()
	return &ConversationBuilder{
		convo: &conversation.Conversation{
			ID:                    1,
			InterviewID:           1,
			CurrentTopic:          1,
			CurrentSubtopic:       "None",
			CurrentQuestionNumber: 2,
			CreatedAt:             now,
			UpdatedAt:             now,
			Topics:                make(map[int]conversation.Topic),
		},
	}
}

func (b *ConversationBuilder) WithTopic(name string, id int) *ConversationBuilder {
	b.convo.Topics[id] = conversation.Topic{
		ID:   id,
		Name: name,
	}
	return b
}

func (b *ConversationBuilder) WithQuestion(topicID, questionNumber int, prompt string) *ConversationBuilder {

	messages := []conversation.Message{}
	topic := b.convo.Topics[topicID]
	topic.ConversationID = b.convo.ID
	if topic.Questions == nil {
		questions := make(map[int]*conversation.Question)
		topic.Questions = questions
	}

	topic.Questions[questionNumber] = &conversation.Question{
		ConversationID: b.convo.ID,
		TopicID:        topicID,
		QuestionNumber: questionNumber,
		Prompt:         prompt,
		CreatedAt:      time.Now().UTC(),
		Messages:       messages,
	}
	b.convo.Topics[topicID] = topic
	return b
}

func (b *ConversationBuilder) WithMessage(topicID, questionNumber int, message []conversation.Message) *ConversationBuilder {
	b.convo.Topics[topicID].Questions[questionNumber].Messages = append(b.convo.Topics[topicID].Questions[questionNumber].Messages, message...)

	return b
}

func (b *ConversationBuilder) WithCurrents(currentTopic, currentQuestionNumber int, currentSubtopic string) *ConversationBuilder {
	b.convo.CurrentTopic = currentTopic
	b.convo.CurrentSubtopic = currentSubtopic
	b.convo.CurrentQuestionNumber = currentQuestionNumber

	return b
}

func (b *ConversationBuilder) Build() *conversation.Conversation {
	return b.convo
}

func NewCreatedConversationMock() *conversation.Conversation {
	builder := NewConversationBuilder()
	builder.WithTopic("Introduction", 1).
		WithQuestion(1, 1, "Question1").
		WithMessage(1, 1, mocks.MessagesCreatedConversationT1Q1).
		WithQuestion(1, 2, "Question2").
		WithMessage(1, 2, mocks.MessagesCreatedConversationT1Q2).
		WithTopic("Coding", 2).
		WithTopic("System Design", 3).
		WithTopic("Databases and Data Management", 4).
		WithTopic("Behavioral", 5).
		WithTopic("General Backend Knowledge", 6)

	return builder.Build()
}

func NewAppendedConversationMock() *conversation.Conversation {
	builder := NewConversationBuilder()
	builder.WithQuestion(2, 1, "Can you write me a func to reverse a string?").
		WithMessage(2, 1, mocks.MessagesAppendedConversationT2Q1)
	return builder.Build()
}
