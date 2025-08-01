package mocks

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/conversation"
)

var now = time.Now().UTC()

const (
	ScenarioInterview  = "interview"
	ScenarioCreated    = "created"
	ScenarioAppended   = "appended"
	ScenarioIsFinished = "finished"
)

var responseFixtures = map[string]*chatgpt.ChatGPTResponse{
	ScenarioInterview: {
		Topic:        "None",
		Subtopic:     "None",
		Question:     "None",
		Score:        0,
		Feedback:     "None",
		NextQuestion: "Question1",
		NextTopic:    "Introduction",
		NextSubtopic: "Subtopic1",
	},
	ScenarioCreated: {
		Topic:        "Introduction",
		Subtopic:     "Subtopic1",
		Question:     "Question1",
		Score:        10,
		Feedback:     "Feedback1",
		NextQuestion: "Question2",
		NextTopic:    "Introduction",
		NextSubtopic: "Subtopic2",
	},
	ScenarioAppended: {
		Topic:        "Introduction",
		Subtopic:     "Subtopic2",
		Question:     "Question2",
		Score:        10,
		Feedback:     "Feedback2",
		NextQuestion: "Question1",
		NextTopic:    "Coding",
		NextSubtopic: "Subtopic1",
	},
	ScenarioIsFinished: {
		Topic:        "General Backend Knowledge",
		Subtopic:     "Subtopic2",
		Question:     "Question2",
		Score:        10,
		Feedback:     "Feedback2",
		NextQuestion: "none",
		NextTopic:    "none",
		NextSubtopic: "none",
	},
}

type MockOpenAIClient struct {
	Scenario string
}

func (m *MockOpenAIClient) GetChatGPTResponse(prompt string) (*chatgpt.ChatGPTResponse, error) {
	return responseFixtures[ScenarioInterview], nil
}

func (m *MockOpenAIClient) GetChatGPTResponseConversation(_ []map[string]string) (*chatgpt.ChatGPTResponse, error) {
	resp, ok := responseFixtures[m.Scenario]
	// DEBUG
	fmt.Printf("resp GetChatGPTResponseConversation: %v\n\n\n", resp)
	if !ok {
		return nil, fmt.Errorf("invalid scenario: %s", m.Scenario)
	}
	return resp, nil
}

func (m *MockOpenAIClient) GetChatGPT35Response(prompt string) (*chatgpt.ChatGPTResponse, error) {
	return &chatgpt.ChatGPTResponse{}, nil
}

func (m *MockOpenAIClient) ExtractJDInput(jd string) (*chatgpt.JDParsedOutput, error) {
	return &chatgpt.JDParsedOutput{}, nil
}

func (m *MockOpenAIClient) ExtractJDSummary(jdInput *chatgpt.JDParsedOutput) (string, error) {
	return "", nil
}

func MarshalAndString(r *chatgpt.ChatGPTResponse) string {
	b, err := json.Marshal(r)
	if err != nil {
		log.Fatalf("Marshal failed: %v", err)
	}
	return string(b)
}

func GetMockMessages(key string) []conversation.Message {
	switch key {
	case "t1q1":
		return []conversation.Message{
			conversation.NewMessage(1, 1, 1, conversation.System, chatgpt.BuildPrompt([]string{}, "Introduction", 1, "")),
			conversation.NewMessage(1, 1, 1, conversation.Interviewer, "Question1"),
			conversation.NewMessage(1, 1, 1, conversation.User, "T1Q1A1"),
		}
	case "t1q2":
		return []conversation.Message{
			conversation.NewMessage(1, 1, 2, conversation.Interviewer, MarshalAndString(responseFixtures[ScenarioCreated])),
		}
	case "t1q2a2":
		return []conversation.Message{
			conversation.NewMessage(1, 1, 2, conversation.User, "T1Q2A2"),
		}
	case "t2q1":
		return []conversation.Message{
			conversation.NewMessage(1, 2, 1, conversation.Interviewer, MarshalAndString(responseFixtures[ScenarioAppended])),
		}
	case "t2q1a1":
		return []conversation.Message{
			conversation.NewMessage(1, 2, 1, conversation.User, "T2Q1A1"),
		}
	case "t2q2":
		return []conversation.Message{
			conversation.NewMessage(1, 2, 2, conversation.Interviewer, MarshalAndString(responseFixtures[ScenarioIsFinished])),
		}
	case "t2q2a2":
		return []conversation.Message{
			conversation.NewMessage(1, 2, 2, conversation.User, "T2Q2A2"),
		}
	case "t2q2a2Finished":
		return []conversation.Message{
			conversation.NewMessage(1, 2, 2, conversation.Interviewer, MarshalAndString(responseFixtures[ScenarioIsFinished])),
		}
	default:
		return nil
	}
}
