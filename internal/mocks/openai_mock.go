package mocks

import (
	"encoding/json"
	"log"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/conversation"
)

var (
	responseConversationCreated     *chatgpt.ChatGPTResponse
	responseConversationMockCreated string
	CreatedConversationMock         *conversation.Conversation
	now                             = time.Now().UTC()
	responseInterview               = &chatgpt.ChatGPTResponse{
		Topic:        "None",
		Subtopic:     "None",
		Question:     "None",
		Score:        0,
		Feedback:     "None",
		NextQuestion: "Tell me a little bit about your work history.",
		NextTopic:    "Introduction",
		NextSubtopic: "General Background",
		CreatedAt:    now,
	}
)

func init() {
	responseConversationCreated = &chatgpt.ChatGPTResponse{
		Topic:        "Introduction",
		Subtopic:     "General Background",
		Question:     "Tell me a little bit about your work history",
		Score:        10,
		Feedback:     "Sounds like you have a good deal of problem solving experience.",
		NextQuestion: "Can you tell me about your most recent backend project?",
		NextTopic:    "Introduction",
		NextSubtopic: "General Engineering Experience",
		CreatedAt:    now,
	}

	responseConversationMarshal, err := json.Marshal(responseConversationCreated)
	if err != nil {
		log.Fatalf("MarshalResponses failed: %v", err)
	}

	responseConversationMockCreated = string(responseConversationMarshal)

	CreatedConversationMock = &conversation.Conversation{
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
								Content:        responseConversationMockCreated,
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

}

type MockOpenAIClient struct{}

func (m *MockOpenAIClient) GetChatGPTResponseInterview(prompt string) (*chatgpt.ChatGPTResponse, error) {
	return responseInterview, nil
}

func (m *MockOpenAIClient) GetChatGPTResponseConversation(conversationHistory []map[string]string) (*chatgpt.ChatGPTResponse, error) {
	return responseConversationCreated, nil
}
