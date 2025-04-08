package testutil

import (
	"time"

	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/handlers"
	"github.com/michaelboegner/interviewer/internal/mocks"
)

type ConversationBuilder struct {
	Convo *conversation.Conversation
}

func NewConversationBuilder() *ConversationBuilder {
	now := time.Now().UTC()
	return &ConversationBuilder{
		Convo: &conversation.Conversation{
			ID:                    1,
			InterviewID:           1,
			CurrentTopic:          1,
			CurrentSubtopic:       "Subtopic2",
			CurrentQuestionNumber: 2,
			CreatedAt:             now,
			UpdatedAt:             now,
			Topics:                make(map[int]conversation.Topic),
		},
	}
}

func (b *ConversationBuilder) WithTopic(name string, id int) *ConversationBuilder {
	b.Convo.Topics[id] = conversation.Topic{
		ID:   id,
		Name: name,
	}
	return b
}

func (b *ConversationBuilder) WithQuestion(topicID, questionNumber int, prompt string) *ConversationBuilder {

	messages := []conversation.Message{}
	topic := b.Convo.Topics[topicID]
	topic.ConversationID = b.Convo.ID
	if topic.Questions == nil {
		questions := make(map[int]*conversation.Question)
		topic.Questions = questions
	}

	topic.Questions[questionNumber] = &conversation.Question{
		ConversationID: b.Convo.ID,
		TopicID:        topicID,
		QuestionNumber: questionNumber,
		Prompt:         prompt,
		CreatedAt:      time.Now().UTC(),
		Messages:       messages,
	}
	b.Convo.Topics[topicID] = topic
	return b
}

func (b *ConversationBuilder) WithMessage(topicID, questionNumber int, messages []conversation.Message) *ConversationBuilder {
	b.Convo.Topics[topicID].Questions[questionNumber].Messages = append(b.Convo.Topics[topicID].Questions[questionNumber].Messages, messages...)

	return b
}

func (b *ConversationBuilder) WithCurrents(currentTopic, currentQuestionNumber int, currentSubtopic string) *ConversationBuilder {
	b.Convo.CurrentTopic = currentTopic
	b.Convo.CurrentSubtopic = currentSubtopic
	b.Convo.CurrentQuestionNumber = currentQuestionNumber

	return b
}

func (b *ConversationBuilder) NewCreatedConversationMock() func() handlers.ReturnVals {
	return func() handlers.ReturnVals {
		b.WithTopic("Introduction", 1).
			WithQuestion(1, 1, "Question1").
			WithMessage(1, 1, mocks.MessagesCreatedConversationT1Q1).
			WithQuestion(1, 2, "Question2").
			WithMessage(1, 2, mocks.MessagesCreatedConversationT1Q2).
			WithTopic("Coding", 2).
			WithTopic("System Design", 3).
			WithTopic("Databases and Data Management", 4).
			WithTopic("Behavioral", 5).
			WithTopic("General Backend Knowledge", 6)
		return handlers.ReturnVals{Conversation: b.Convo}
	}
}

func (b *ConversationBuilder) NewAppendedConversationMock() func() handlers.ReturnVals {
	return func() handlers.ReturnVals {
		b.WithCurrents(2, 1, "Subtopic1").
			WithMessage(1, 2, mocks.MessagesCreatedConversationT1Q2A2).
			WithQuestion(2, 1, "Question1").
			WithMessage(2, 1, mocks.MessagesAppendedConversationT2Q1)
		return handlers.ReturnVals{Conversation: b.Convo}
	}
}

func (b *ConversationBuilder) NewIsFinishedConversationMock() *conversation.Conversation {
	b.WithCurrents(0, 0, "Finished").
		WithQuestion(6, 1, "Question1").
		WithMessage(6, 1, mocks.MessagesAppendedConversationT6Q1A1)

	return b.Convo
}
