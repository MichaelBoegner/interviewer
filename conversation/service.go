package conversation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func CheckForConversation(repo ConversationRepo, interviewID int) bool {
	return repo.CheckForConversation(interviewID)
}

func CreateConversation(repo ConversationRepo, interviewID int, firstQuestion string, messageResponse *Message) (*Conversation, error) {
	now := time.Now()
	conversation := &Conversation{
		InterviewID: interviewID,
		Topics:      PredefinedTopics,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	conversationID, err := repo.CreateConversation(conversation)
	if err != nil {
		return nil, err
	}

	conversation.ID = conversationID

	questionID, err := repo.CreateQuestion(conversation)
	if err != nil {
		return nil, err
	}

	topic := conversation.Topics[1]
	topic.ConversationID = conversationID
	topic.Questions = make(map[int]Question)

	question := topic.Questions[1]
	question.ID = questionID
	question.QuestionNumber = 1
	question.Prompt = firstQuestion

	messageFirst := &Message{
		QuestionID: questionID,
		Author:     "Interviewer",
		Content:    firstQuestion,
		CreatedAt:  time.Now(),
	}

	messageResponse.ID = 2
	messageResponse.QuestionID = questionID
	messageResponse.CreatedAt = time.Now()

	question.Messages = make([]Message, 0)
	question.Messages = append(question.Messages, *messageFirst)
	question.Messages = append(question.Messages, *messageResponse)

	err = repo.CreateMessages(conversation, question.Messages)
	if err != nil {
		return nil, err
	}

	conversation.Topics[1] = topic
	conversation.Topics[1].Questions[1] = question

	return conversation, nil
}

func AppendConversation(repo ConversationRepo, conversation *Conversation, message *Message, conversationID, topicID, questionID int) (*Conversation, error) {
	if conversation.ID != conversationID {
		return nil, errors.New("conversation_id doesn't match with current interview")
	}

	messageID, err := repo.AddMessage(questionID, message)
	if err != nil {
		return nil, err
	}

	messageToAppend := &Message{
		ID:         messageID,
		QuestionID: questionID,
		Author:     message.Author,
		Content:    message.Content,
		CreatedAt:  time.Now(),
	}

	topic := conversation.Topics[topicID]
	for _, question := range topic.Questions {
		if question.ID == questionID {
			question.Messages = append(question.Messages, *messageToAppend)
		}
	}

	conversation.Topics[topicID] = topic

	return conversation, nil
}

func GetConversation(repo ConversationRepo, interviewID int) (*Conversation, error) {
	conversation, err := repo.GetConversation(interviewID)
	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func getNextQuestion(conversation *Conversation) (string, error) {
	ctx := context.Background()
	apiKey := os.Getenv("OPENAI_API_KEY")

	systemMessage := map[string]string{
		"role":    "system",
		"content": "You are conducting a technical interview for a backend development position. The candidate is at a junior to mid-level skill level. Continue the interview by asking relevant technical questions that assess backend development skills.",
	}

	conversationHistory, err := getConversationHistory(conversation, topicID, questionNumber)
	if err != nil {
		return "", err
	}

	completeConversation := append([]map[string]string{systemMessage}, conversationHistory...)

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       "gpt-4",
		"messages":    []map[string]string{completeConversation},
		"max_tokens":  150,
		"temperature": 0.7,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	responseChan := make(chan string)
	errorChan := make(chan error)
	go func() {
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errorChan <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errorChan <- fmt.Errorf("API call failed with status code: %d", resp.StatusCode)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errorChan <- err
			return
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			errorChan <- err
			return
		}

		choices := result["choices"].([]interface{})
		if len(choices) == 0 {
			errorChan <- fmt.Errorf("no question generated")
			return
		}

		firstQuestion := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
		responseChan <- firstQuestion
	}()

	select {
	case firstQuestion := <-responseChan:
		response := firstQuestion
		return response, nil

	case err := <-errorChan:
		return "", err
	}
}

func getConversationHistory(conversation *Conversation, topicID, questionNumber int) ([]map[string]string, error) {
	chatGPTConversationArray := make([]map[string]string, 0)

	for _, message := range conversation.Topics[topicID].Questions[questionNumber].Messages {
		conversationMap := make(map[string]string)
		role := string(message.Author)
		content := message.Content
		conversationMap[role] = content

		chatGPTConversationArray = append(chatGPTConversationArray, conversationMap)
	}

	return chatGPTConversationArray, nil
}
