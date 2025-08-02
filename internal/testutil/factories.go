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
			Topics:                conversation.ClonePredefinedTopics(),
		},
	}
}

func (b *ConversationBuilder) WithTopic(name string, id int) *ConversationBuilder {
	b.Convo.Topics[id] = &conversation.Topic{
		ID:        id,
		Name:      name,
		Questions: map[int]*conversation.Question{},
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
		b := NewConversationBuilder()
		b.WithTopic("Introduction", 1).
			WithQuestion(1, 1, "Question1").
			WithMessage(1, 1, mocks.GetMockMessages("t1q1")).
			WithQuestion(1, 2, "Question2").
			WithMessage(1, 2, mocks.GetMockMessages("t1q2"))

		return handlers.ReturnVals{Conversation: b.Convo}
	}
}

func NewAppendedConversationMock() func() handlers.ReturnVals {
	return func() handlers.ReturnVals {
		b := NewConversationBuilder()
		b.WithTopic("Introduction", 1).
			WithCurrents(2, 1, "Subtopic1").
			WithQuestion(1, 1, "Question1").
			WithMessage(1, 1, mocks.GetMockMessages("t1q1")).
			WithQuestion(1, 2, "Question2").
			WithMessage(1, 2, mocks.GetMockMessages("t1q2")).
			WithMessage(1, 2, mocks.GetMockMessages("t1q2a2")).
			WithQuestion(2, 1, "Question1").
			WithMessage(2, 1, mocks.GetMockMessages("t2q1"))

		return handlers.ReturnVals{Conversation: b.Convo}
	}
}

func NewIsFinishedConversationMock() func() handlers.ReturnVals {
	return func() handlers.ReturnVals {
		b := NewConversationBuilder()
		b.WithCurrents(0, 0, "finished").
			WithQuestion(1, 1, "Question1").
			WithMessage(1, 1, mocks.GetMockMessages("t1q1")).
			WithQuestion(1, 2, "Question2").
			WithMessage(1, 2, mocks.GetMockMessages("t1q2")).
			WithMessage(1, 2, mocks.GetMockMessages("t1q2a2")).
			WithQuestion(2, 1, "Question1").
			WithMessage(2, 1, mocks.GetMockMessages("t2q1")).
			WithMessage(2, 1, mocks.GetMockMessages("t2q1a1")).
			WithQuestion(2, 2, "Question2").
			WithMessage(2, 2, mocks.GetMockMessages("t2q2")).
			WithMessage(2, 2, mocks.GetMockMessages("t2q2a2")).
			WithMessage(2, 2, mocks.GetMockMessages("t2q2a2Finished"))

		return handlers.ReturnVals{Conversation: b.Convo}
	}
}
