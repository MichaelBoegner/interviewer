package testutil

import (
	"time"

	"github.com/michaelboegner/interviewer/conversation"
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
			CurrentQuestionNumber: 1,
			CreatedAt:             now,
			UpdatedAt:             now,
			Topics:                make(map[int]conversation.Topic),
		},
	}
}

func (b *ConversationBuilder) WithTopic(name string, id int) *ConversationBuilder {
	b.convo.Topics[id] = conversation.Topic{
		ID:             id,
		ConversationID: b.convo.ID,
		Name:           name,
		Questions:      make(map[int]*conversation.Question),
	}
	return b
}

func (b *ConversationBuilder) WithQuestion(topicID, qNum int, prompt string, messages []conversation.Message) *ConversationBuilder {
	topic := b.convo.Topics[topicID]
	topic.Questions[qNum] = &conversation.Question{
		ConversationID: b.convo.ID,
		TopicID:        topicID,
		QuestionNumber: qNum,
		Prompt:         prompt,
		CreatedAt:      time.Now().UTC(),
		Messages:       messages,
	}
	b.convo.Topics[topicID] = topic
	return b
}

func (b *ConversationBuilder) Build() *conversation.Conversation {
	return b.convo
}
