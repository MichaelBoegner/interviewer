package mocks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/conversation"
)

// TODO: This is a janky implementation with no long term scalability, at least not without a high amount of redundancies.
// Message creation should be automated in some manner and the mock chatGPT response logic doesn't allow for easy scalability.
// Also, since we are only mocking create, append, and isFinished, the tested convo structure isn't realistic, though it still proves correct functionality.
// Will come back to this after finishing overall testing structure, CI/CD, and AWS deployment.
var (
	responseConversationMockCreated    string
	responseConversationMockAppended   string
	CreatedConversationMock            *conversation.Conversation
	MessagesCreatedConversationT1Q1    []conversation.Message
	MessagesCreatedConversationT1Q2    []conversation.Message
	MessagesCreatedConversationT1Q2A2  []conversation.Message
	MessagesAppendedConversationT2Q1   []conversation.Message
	MessagesAppendedConversationT2Q1A1 []conversation.Message
	MessagesAppendedConversationT2Q2   []conversation.Message
	MessagesAppendedConversationT6Q1A1 []conversation.Message

	now               = time.Now().UTC()
	responseInterview = &chatgpt.ChatGPTResponse{
		Topic:        "None",
		Subtopic:     "None",
		Question:     "None",
		Score:        0,
		Feedback:     "None",
		NextQuestion: "Question1",
		NextTopic:    "Introduction",
		NextSubtopic: "Subtopic1",
	}
	responseConversationCreated = &chatgpt.ChatGPTResponse{
		Topic:        "Introduction",
		Subtopic:     "Subtopic1",
		Question:     "Question1",
		Score:        10,
		Feedback:     "Feedback1",
		NextQuestion: "Question2",
		NextTopic:    "Introduction",
		NextSubtopic: "Subtopic2",
	}
	responseConversationAppended = &chatgpt.ChatGPTResponse{
		Topic:        "Introduction",
		Subtopic:     "Subtopic2",
		Question:     "Question2",
		Score:        10,
		Feedback:     "Feedback2",
		NextQuestion: "Question1",
		NextTopic:    "Coding",
		NextSubtopic: "Subtopic1",
	}

	responseConversationIsFinished = &chatgpt.ChatGPTResponse{
		Topic:        "General Backend Knowledge",
		Subtopic:     "Subtopic1",
		Question:     "Question1",
		Score:        10,
		Feedback:     "Feedback1",
		NextQuestion: "Question2",
		NextTopic:    "General Backend Knowledge",
		NextSubtopic: "Subtopic2",
	}
)

func init() {
	responseConversationMockCreated, err := MarshalAndString(responseConversationCreated)
	if err != nil {
		log.Fatalf("MarshalAndString failed: %v", err)
	}

	responseConversationMockAppended, err := MarshalAndString(responseConversationAppended)
	if err != nil {
		log.Fatalf("MarshalAndString failed: %v", err)
	}

	responseConversationMockIsFinished, err := MarshalAndString(responseConversationIsFinished)
	if err != nil {
		log.Fatalf("MarshalAndString failed: %v", err)
	}

	MessagesCreatedConversationT1Q1 = []conversation.Message{
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
			Content:        "Question1",
			CreatedAt:      now,
		},
		{
			ConversationID: 1,
			TopicID:        1,
			QuestionNumber: 1,
			Author:         "user",
			Content:        "Answer1",
			CreatedAt:      now,
		},
	}

	MessagesCreatedConversationT1Q2 = []conversation.Message{
		{
			ConversationID: 1,
			TopicID:        1,
			QuestionNumber: 2,
			Author:         "interviewer",
			Content:        responseConversationMockCreated,
			CreatedAt:      now,
		},
	}

	MessagesCreatedConversationT1Q2A2 = []conversation.Message{
		{
			ConversationID: 1,
			TopicID:        1,
			QuestionNumber: 2,
			Author:         "user",
			Content:        "Answer2",
			CreatedAt:      now,
		},
	}

	MessagesAppendedConversationT2Q1 = []conversation.Message{
		{
			ConversationID: 1,
			TopicID:        2,
			QuestionNumber: 1,
			Author:         "interviewer",
			Content:        responseConversationMockAppended,
			CreatedAt:      now,
		},
	}

	MessagesAppendedConversationT2Q1A1 = []conversation.Message{
		{
			ConversationID: 1,
			TopicID:        2,
			QuestionNumber: 1,
			Author:         "user",
			Content:        "Answer1",
			CreatedAt:      now,
		},
	}

	MessagesAppendedConversationT2Q2 = []conversation.Message{
		{
			ConversationID: 1,
			TopicID:        0,
			QuestionNumber: 1,
			Author:         "interviewer",
			Content:        responseConversationMockIsFinished,
			CreatedAt:      now,
		},
	}

	MessagesAppendedConversationT6Q1A1 = []conversation.Message{
		{
			ConversationID: 1,
			TopicID:        6,
			QuestionNumber: 1,
			Author:         "user",
			Content:        "Answer1",
			CreatedAt:      now,
		},
	}

}

type MockOpenAIClient struct {
	APIKey string
}

func (m *MockOpenAIClient) GetChatGPTResponse(prompt string) (*chatgpt.ChatGPTResponse, error) {
	return responseInterview, nil
}

func (m *MockOpenAIClient) GetChatGPTResponseConversation(conversationHistory []map[string]string) (*chatgpt.ChatGPTResponse, error) {
	if len(conversationHistory) == 3 && !strings.Contains(conversationHistory[1]["content"], "Coding") {
		return responseConversationCreated, nil
	} else if len(conversationHistory) == 6 {
		return responseConversationAppended, nil
	}

	return responseConversationIsFinished, nil
}

func (m *MockOpenAIClient) GetChatGPT35Response(prompt string) (*chatgpt.ChatGPTResponse, error) {
	ctx := context.Background()

	var messagesArray []map[string]string
	messagesArray = append(messagesArray, map[string]string{
		"role":    "system",
		"content": prompt,
	})

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       "gpt-3.5-turbo-0125",
		"messages":    messagesArray,
		"max_tokens":  1000,
		"temperature": 0,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("NewRequestWithContext failed: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.APIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, &chatgpt.OpenAIError{StatusCode: resp.StatusCode, Message: errResp.Error.Message}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Unmarshal result err: %v", err)
		return nil, err
	}

	choices := result["choices"].([]interface{})
	if len(choices) == 0 {
		err := errors.New("no question generated")
		return nil, err
	}

	chatGPTResponseRaw := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	var chatGPTResponse chatgpt.ChatGPTResponse
	if err := json.Unmarshal([]byte(chatGPTResponseRaw), &chatGPTResponse); err != nil {
		log.Printf("Unmarshal chatGPTResponse err: %v", err)
		return nil, err
	}

	return &chatGPTResponse, nil
}

func (m *MockOpenAIClient) ExtractJDInput(jd string) (*chatgpt.JDParsedOutput, error) {
	systemPrompt := chatgpt.BuildJDPromptInput(jd)
	response, err := m.GetChatGPT35Response(systemPrompt)
	if err != nil {
		return nil, err
	}

	return &chatgpt.JDParsedOutput{
		Responsibilities: response.Responsibilities,
		Qualifications:   response.Qualifications,
		TechStack:        response.TechStack,
		Level:            response.Level,
	}, nil
}

func (m *MockOpenAIClient) ExtractJDSummary(jdInput *chatgpt.JDParsedOutput) (string, error) {
	jdJSON, err := json.MarshalIndent(jdInput, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JDParsedOutput: %w", err)
	}

	systemPrompt := chatgpt.BuildJDPromptSummary(string(jdJSON))
	response, err := m.GetChatGPT35Response(systemPrompt)
	if err != nil {
		return "", err
	}

	jdSummary := fmt.Sprintf(`### JD Context

- Level: %s
- Domain: %s
- Tech Stack: %s
- Responsibilities: %s
- Qualifications: %s
`,
		response.Level,
		response.Domain,
		strings.Join(response.TechStack, ", "),
		strings.Join(response.Responsibilities, "; "),
		strings.Join(response.Qualifications, "; "),
	)

	return jdSummary, nil
}

func MarshalAndString(chatGPTResponse *chatgpt.ChatGPTResponse) (string, error) {
	chatGPTResponseMarshal, err := json.Marshal(chatGPTResponse)
	if err != nil {
		log.Fatalf("MarshalResponses failed: %v", err)
		return "", nil
	}

	return string(chatGPTResponseMarshal), nil
}
