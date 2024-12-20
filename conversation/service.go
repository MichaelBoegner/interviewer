package conversation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/michaelboegner/interviewer/models"
)

func CheckForConversation(repo ConversationRepo, interviewID int) bool {
	return repo.CheckForConversation(interviewID)
}

func CreateConversation(
	repo ConversationRepo,
	interviewID int,
	prompt string,
	firstQuestion string,
	messageUserResponse *Message) (*Conversation, error) {
	now := time.Now()

	if messageUserResponse == nil {
		log.Printf("messageUserResponse is nil")
		return nil, errors.New("messageUserResponse cannot be nil")
	}

	conversation := &Conversation{
		InterviewID: interviewID,
		Topics:      PredefinedTopics,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	conversationID, err := repo.CreateConversation(conversation)
	if err != nil {
		log.Printf("CreateConversation failing")
		return nil, err
	}
	conversation.ID = conversationID

	questionID, err := repo.CreateQuestion(conversation, firstQuestion)
	if err != nil {
		log.Printf("CreateQuestion failing")
		return nil, err
	}

	topic := conversation.Topics[1]
	topic.ConversationID = conversationID

	messagePrompt := newMessage(1, questionID, System, prompt)
	messageFirstQuestion := newMessage(2, questionID, Interviewer, firstQuestion)

	messageUserResponse.ID = 3
	messageUserResponse.QuestionID = questionID
	messageUserResponse.CreatedAt = time.Now()

	question := &Question{
		ID:             questionID,
		QuestionNumber: 1,
		ConversationID: conversationID,
		TopicID:        1,
		Prompt:         firstQuestion,
		Messages: []Message{
			*messagePrompt,
			*messageFirstQuestion,
			*messageUserResponse,
		},
		CreatedAt: time.Now(),
	}

	topic.Questions = make(map[int]*Question)
	topic.Questions[1] = question

	conversation.Topics[1] = topic

	chatGPTResponse, err := getNextQuestion(conversation, 1, 1)
	if err != nil {
		log.Printf("getNextQuestion failing")
		return nil, err
	}

	chatGPTResponseString, err := ChatGPTResponseToString(chatGPTResponse)
	if err != nil {
		log.Printf("Marshalled response err: %v", err)
		return nil, err
	}

	topic = conversation.Topics[1]
	topic.Questions[1].Messages = append(topic.Questions[1].Messages, *newMessage(3, questionID, Interviewer, chatGPTResponseString))
	conversation.Topics[1] = topic

	err = repo.CreateMessages(conversation, question.Messages)
	if err != nil {
		log.Printf("repo.CreateMessages failing")
		return nil, err
	}

	return conversation, nil
}

func AppendConversation(
	repo ConversationRepo,
	conversation *Conversation,
	message *Message,
	conversationID, topicID, questionID, questionNumber int,
	prompt string) (*Conversation, error) {
	if conversation.ID != conversationID {
		return nil, errors.New("conversation_id doesn't match with current interview")
	}

	messageID, err := repo.AddMessage(questionID, message)
	if err != nil {
		return nil, err
	}

	messageUser := newMessage(messageID, questionID, message.Author, message.Content)

	messages := conversation.Topics[topicID].Questions[questionNumber].Messages
	messages = append(messages, *messageUser)
	conversation.Topics[topicID].Questions[questionNumber].Messages = messages

	chatGPTResponse, err := getNextQuestion(conversation, topicID, questionNumber)
	if err != nil {
		log.Printf("getNextQuestion failing")
		return nil, err
	}

	moveToNewTopic, _, isFinished := getConversationState(chatGPTResponse)

	chatGPTResponseString, err := ChatGPTResponseToString(chatGPTResponse)
	if err != nil {
		log.Printf("Marshalled response err: %v", err)
		return nil, err
	}

	//Debug printing
	data, _ := json.MarshalIndent(chatGPTResponse, "", "  ")
	fmt.Println(string(data))
	fmt.Printf("movetoNewTopic: %v\n", moveToNewTopic)

	if isFinished {
		return conversation, nil
	}

	if moveToNewTopic {
		fmt.Printf("moveToNewTopic: %v\n", moveToNewTopic)
		topicID += 1

		topic := conversation.Topics[topicID]
		topic.ConversationID = conversationID

		messagePrompt := newMessage(1, questionID, System, prompt)
		messageFirstQuestion := newMessage(2, questionID, Interviewer, chatGPTResponseString)

		question := &Question{
			ID:             questionID,
			QuestionNumber: 1,
			ConversationID: conversationID,
			TopicID:        topicID,
			Prompt:         chatGPTResponse.NextQuestion,
			Messages: []Message{
				*messagePrompt,
				*messageFirstQuestion,
			},
			CreatedAt: time.Now(),
		}

		topic.Questions = make(map[int]*Question)
		topic.Questions[1] = question

		conversation.Topics[topicID] = topic

		_, err = repo.AddMessage(questionID, messagePrompt)
		if err != nil {
			return nil, err
		}

		return conversation, nil

	}

	messageNextQuestionID := messageID + 1
	messageNextQuestion := newMessage(messageNextQuestionID, questionID, Interviewer, chatGPTResponseString)

	messages = conversation.Topics[topicID].Questions[questionNumber].Messages
	messages = append(messages, *messageNextQuestion)
	conversation.Topics[topicID].Questions[questionNumber].Messages = messages

	_, err = repo.AddMessage(questionID, messageNextQuestion)
	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func GetConversation(repo ConversationRepo, interviewID, questionID int) (*Conversation, error) {
	conversation, err := repo.GetConversation(interviewID)
	if err != nil {
		return nil, err
	}

	conversation.Topics = PredefinedTopics
	topic := conversation.Topics[1]
	topic.ConversationID = conversation.ID
	topic.Questions = make(map[int]*Question)

	questionReturned, err := repo.GetQuestion(conversation)
	if err != nil {
		return nil, err
	}

	topic.Questions[1] = questionReturned

	question := topic.Questions[1]
	question.Messages = make([]Message, 0)

	messagesReturned, err := repo.GetMessages(questionID)
	if err != nil {
		log.Printf("repo.GetMessages failed: %v\n", err)
		return nil, err
	}

	question.Messages = append(question.Messages, messagesReturned...)
	conversation.Topics[1] = topic
	conversation.Topics[1].Questions[1] = question

	return conversation, nil
}

func getNextQuestion(conversation *Conversation, topicID, questionNumber int) (*models.ChatGPTResponse, error) {
	ctx := context.Background()
	apiKey := os.Getenv("OPENAI_API_KEY")

	conversationHistory, err := getConversationHistory(conversation, topicID, questionNumber)
	if err != nil {
		return nil, err
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       "gpt-4",
		"messages":    conversationHistory,
		"max_tokens":  150,
		"temperature": 0.7,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("NewRequestWithContext failing")
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("API call failed with status code: %d, response: %s", resp.StatusCode, body)
		return nil, fmt.Errorf("API call failed with status code: %d", resp.StatusCode)
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

	chatGPTResponseResponse := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	fmt.Printf("chatGPTResponseResponse: %v", chatGPTResponseResponse)
	var chatGPTResponse models.ChatGPTResponse
	if err := json.Unmarshal([]byte(chatGPTResponseResponse), &chatGPTResponse); err != nil {
		log.Printf("Unmarshal chatGPTResponse err: %v", err)
		return nil, err
	}

	return &chatGPTResponse, nil

}

func getConversationHistory(conversation *Conversation, topicID, questionNumber int) ([]map[string]string, error) {
	chatGPTConversationArray := make([]map[string]string, 0)

	for _, message := range conversation.Topics[topicID].Questions[questionNumber].Messages {
		conversationMap := make(map[string]string)
		var role string
		if message.Author == "interviewer" {
			role = "assistant"
		} else {
			role = string(message.Author)
		}
		conversationMap["role"] = role
		content := message.Content
		conversationMap["content"] = content

		chatGPTConversationArray = append(chatGPTConversationArray, conversationMap)
	}

	return chatGPTConversationArray, nil
}

func newMessage(messageID, questionID int, author Author, content string) *Message {
	message := &Message{
		ID:         messageID,
		QuestionID: questionID,
		Author:     author,
		Content:    content,
		CreatedAt:  time.Now(),
	}

	return message
}

func ChatGPTResponseToString(chatGPTResponse *models.ChatGPTResponse) (string, error) {
	chatGPTResponseString, err := json.Marshal(chatGPTResponse)
	if err != nil {
		log.Printf("chatGPTResponseToString failed: %v", err)
		return "", err
	}

	return string(chatGPTResponseString), nil
}

func getConversationState(chatGPTResponse *models.ChatGPTResponse) (bool, bool, bool) {
	isFinished := false
	if chatGPTResponse.NextQuestion == "finished" {
		isFinished = true
	}

	return chatGPTResponse.MoveToNewTopic, chatGPTResponse.MoveToNewSubtopic, isFinished
}
