package mocks

import "github.com/michaelboegner/interviewer/conversation"

var responseConversationForMock, err = MarshalResponses(responseConversation)

var TestCreatedConversation = &conversation.Conversation{
	ID:                    1,
	InterviewID:           1,
	CurrentTopic:          1,
	CurrentSubtopic:       "None",
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
							Author:         "system",
							Content:        TestPrompt,
						},
						{
							ConversationID: 1,
							TopicID:        1,
							QuestionNumber: 1,
							Author:         "interviewer",
							Content:        "Tell me a little bit about your work history.",
						},
						{
							ConversationID: 1,
							TopicID:        1,
							QuestionNumber: 1,
							Author:         "user",
							Content:        "I have been a TSE for 5 years.",
						},
						{
							ConversationID: 1,
							TopicID:        1,
							QuestionNumber: 1,
							Author:         "interviewer",
							Content:        responseConversationForMock,
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
			ID:             4,
			ConversationID: 1,
			Name:           "Databases and Data Management",
			Questions:      map[int]*conversation.Question{},
		},
		5: {
			ID:             5,
			ConversationID: 1,
			Name:           "Behavioral",
			Questions:      map[int]*conversation.Question{},
		},
		6: {
			ID:             6,
			ConversationID: 1,
			Name:           "General Backend Knowledge",
			Questions:      map[int]*conversation.Question{},
		},
	},
}
