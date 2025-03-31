package mocks

import (
	"time"

	"github.com/michaelboegner/interviewer/conversation"
)

var now = time.Now()

var TestCreatedConversation = &conversation.Conversation{
	ID:                    1,
	InterviewID:           1,
	CurrentTopic:          1,
	CurrentSubtopic:       "None",
	CurrentQuestionNumber: 1,
	CreatedAt:             now,
	UpdatedAt:             now,
	Topics: map[int]conversation.Topic{
		1: {
			ID:             1,
			ConversationID: 1,
			Name:           "Introduction",
			Questions: map[int]*conversation.Question{
				1: {
					ConversationID: 1,
					TopicID:        1,
					QuestionNumber: 1,
					Prompt:         "Tell me a little bit about your work history.",
					CreatedAt:      now,
					Messages: []conversation.Message{
						{
							ConversationID: 1,
							TopicID:        1,
							QuestionNumber: 1,
							Author:         "system",
							Content:        TestPrompt,
							CreatedAt:      now,
						},
						{
							ConversationID: 1,
							TopicID:        1,
							QuestionNumber: 1,
							Author:         "interviewer",
							Content:        "Tell me a little bit about your work history.",
							CreatedAt:      now,
						},
						{
							ConversationID: 1,
							TopicID:        1,
							QuestionNumber: 1,
							Author:         "user",
							Content:        "I have been a TSE for 5 years.",
							CreatedAt:      now,
						},
						{
							ConversationID: 1,
							TopicID:        1,
							QuestionNumber: 1,
							Author:         "interviewer",
							Content:        responseConversationMock,
							CreatedAt:      now,
						},
					},
				},
			},
		},
		2: {
			ID:             2,
			ConversationID: 0,
			Name:           "Coding",
			Questions:      nil,
		},
		3: {
			ID:             3,
			ConversationID: 0,
			Name:           "System Design",
			Questions:      nil,
		},
		4: {
			ID:             4,
			ConversationID: 0,
			Name:           "Databases and Data Management",
			Questions:      nil,
		},
		5: {
			ID:             5,
			ConversationID: 0,
			Name:           "Behavioral",
			Questions:      nil,
		},
		6: {
			ID:             6,
			ConversationID: 0,
			Name:           "General Backend Knowledge",
			Questions:      nil,
		},
	},
}
