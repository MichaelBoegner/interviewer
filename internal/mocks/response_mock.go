package mocks

import "github.com/michaelboegner/interviewer/conversation"

var TestCreatedConversation = &conversation.Conversation{
	InterviewID:           1,
	CurrentTopic:          1,
	CurrentSubtopic:       "General Background",
	CurrentQuestionNumber: 1,
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
					Messages: []conversation.Message{
						{
							ConversationID: 1,
							TopicID:        1,
							QuestionNumber: 1,
							Author:         "interviewer",
							Content:        "Tell me a little bit about your work history.",
						},
					},
				},
			},
		},
		2: {
			ID:             2,
			ConversationID: 1,
			Name:           "Coding",
			Questions:      map[int]*conversation.Question{},
		},
		3: {
			ID:             3,
			ConversationID: 1,
			Name:           "System Design",
			Questions:      map[int]*conversation.Question{},
		},
		4: {
			ID:             3,
			ConversationID: 1,
			Name:           "Databases and Data Management",
			Questions:      map[int]*conversation.Question{},
		},
		5: {
			ID:             3,
			ConversationID: 1,
			Name:           "Behavioral",
			Questions:      map[int]*conversation.Question{},
		},
		6: {
			ID:             3,
			ConversationID: 1,
			Name:           "General Backend Knowledge",
			Questions:      map[int]*conversation.Question{},
		},
	},
}
