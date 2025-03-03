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
		InterviewID:           interviewID,
		Topics:                PredefinedTopics,
		CurrentTopic:          1,
		CurrentQuestionNumber: 1,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	conversationID, err := repo.CreateConversation(conversation)
	if err != nil {
		log.Printf("CreateConversation failing: %v", err)
		return nil, err
	}
	conversation.ID = conversationID

	questionNumber, err := repo.CreateQuestion(conversation, firstQuestion)
	if err != nil {
		log.Printf("CreateQuestion failing")
		return nil, err
	}

	topic := conversation.Topics[1]
	topic.ConversationID = conversationID

	messagePrompt := newMessage(conversationID, questionNumber, System, prompt)
	messageFirstQuestion := newMessage(conversationID, questionNumber, Interviewer, firstQuestion)

	messageUserResponse.ConversationID = conversationID
	messageUserResponse.QuestionNumber = questionNumber
	messageUserResponse.CreatedAt = time.Now()

	question := &Question{
		ConversationID: conversationID,
		TopicID:        1,
		QuestionNumber: questionNumber,
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

	chatGPTResponse, err := getNextQuestion(conversation)
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
	topic.Questions[1].Messages = append(topic.Questions[1].Messages, *newMessage(conversationID, questionNumber, Interviewer, chatGPTResponseString))
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

	_, err := repo.AddMessage(conversationID, conversation.CurrentTopic, questionNumber, message)
	if err != nil {
		return nil, err
	}

	messageUser := newMessage(conversationID, questionNumber, message.Author, message.Content)

	messages := conversation.Topics[topicID].Questions[questionNumber].Messages
	messages = append(messages, *messageUser)
	conversation.Topics[topicID].Questions[questionNumber].Messages = messages

	chatGPTResponse, err := getNextQuestion(conversation)
	if err != nil {
		log.Printf("getNextQuestion err: %v", err)
		return nil, err
	}

	moveToNewTopic, _, isFinished := getConversationState(chatGPTResponse, conversation)

	chatGPTResponseString, err := ChatGPTResponseToString(chatGPTResponse)
	if err != nil {
		log.Printf("Marshalled response err: %v", err)
		return nil, err
	}

	if isFinished {
		return conversation, nil
	}

	if moveToNewTopic {
		fmt.Printf("\n\n\nmoveToNewTopic: %v\n", moveToNewTopic)
		fmt.Printf("conversation.CurrentTopic: %v\n\n\n", conversation.CurrentTopic)
		nextTopicID := topicID + 1
		nextQuestionNumber := 1
		_, err := repo.UpdateConversationCurrents(nextTopicID, nextQuestionNumber, conversationID)
		if err != nil {
			log.Printf("UpdateConversationTopic error: %v", err)
			return nil, err
		}

		topic := conversation.Topics[nextTopicID]
		topic.ConversationID = conversationID

		messageFirstQuestion := newMessage(conversationID, nextQuestionNumber, Interviewer, chatGPTResponseString)

		question := &Question{
			ConversationID: conversationID,
			TopicID:        nextTopicID,
			QuestionNumber: nextQuestionNumber,
			Prompt:         chatGPTResponse.NextQuestion,
			Messages: []Message{
				*messageFirstQuestion,
			},
			CreatedAt: time.Now(),
		}

		topic.Questions = make(map[int]*Question)
		topic.Questions[nextQuestionNumber] = question

		conversation.Topics[nextTopicID] = topic

		_, err = repo.AddQuestion(question)
		if err != nil {
			log.Printf("AddQuestion in AppendConversation err: %v", err)
		}

		_, err = repo.AddMessage(conversationID, conversation.CurrentTopic, questionNumber, messageFirstQuestion)
		if err != nil {
			return nil, err
		}

		return conversation, nil
	}

	messageNextQuestion := newMessage(conversationID, questionNumber, Interviewer, chatGPTResponseString)

	messages = conversation.Topics[topicID].Questions[questionNumber].Messages
	messages = append(messages, *messageNextQuestion)
	conversation.Topics[topicID].Questions[questionNumber].Messages = messages

	_, err = repo.AddMessage(conversationID, conversation.CurrentTopic, questionNumber, messageNextQuestion)
	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func GetConversation(repo ConversationRepo, interviewID int) (*Conversation, error) {
	conversation, err := repo.GetConversation(interviewID)
	if err != nil {
		return nil, err
	}

	conversation.Topics = PredefinedTopics

	questionsReturned, err := repo.GetQuestions(conversation)
	if err != nil {
		return nil, err
	}

	//PRINT QUESTIONS RETURNED
	for i, q := range questionsReturned {
		fmt.Printf("\nquestionsReturned[%d]: \n%+v\n", i, *q)
	}

	for topicID := 1; topicID <= conversation.CurrentTopic; topicID++ {
		topic := conversation.Topics[topicID]
		topic.ConversationID = conversation.ID
		topic.Questions = make(map[int]*Question)

		for questionNumber := 1; questionNumber <= conversation.CurrentQuestionNumber; questionNumber++ {
			if questionsReturned[questionNumber-1].TopicID != topicID {
				break
			}

			topic.Questions[questionNumber] = questionsReturned[questionNumber-1]

			question := topic.Questions[questionNumber]
			question.Messages = make([]Message, 0)

			messagesReturned, err := repo.GetMessages(conversation.ID, conversation.CurrentTopic, questionNumber)
			if err != nil {
				log.Printf("repo.GetMessages failed: %v\n", err)
				return nil, err
			}

			question.Messages = append(question.Messages, messagesReturned...)
			conversation.Topics[topicID] = topic
			conversation.Topics[topicID].Questions[questionNumber] = question
		}
	}

	return conversation, nil
}

func getNextQuestion(conversation *Conversation) (*models.ChatGPTResponse, error) {
	ctx := context.Background()
	apiKey := os.Getenv("OPENAI_API_KEY")

	conversationHistory, err := getConversationHistory(conversation)
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

	chatGPTResponseRaw := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	fmt.Printf("\n\nchatGPTResponseRaw: %v\n\n", chatGPTResponseRaw)

	var chatGPTResponse models.ChatGPTResponse
	if err := json.Unmarshal([]byte(chatGPTResponseRaw), &chatGPTResponse); err != nil {
		log.Printf("Unmarshal chatGPTResponse err: %v", err)
		return nil, err
	}

	return &chatGPTResponse, nil

}

func getConversationHistory(conversation *Conversation) ([]map[string]string, error) {
	chatGPTConversationArray := make([]map[string]string, 0)

	for _, topic := range conversation.Topics {
		for _, question := range topic.Questions {
			for _, message := range question.Messages {
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
		}
	}

	return chatGPTConversationArray, nil
}

func newMessage(conversationID, currentQuestionNumber int, author Author, content string) *Message {
	message := &Message{
		ConversationID: conversationID,
		QuestionNumber: currentQuestionNumber,
		Author:         author,
		Content:        content,
		CreatedAt:      time.Now(),
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

func getConversationState(chatGPTResponse *models.ChatGPTResponse, conversation *Conversation) (bool, bool, bool) {
	isFinished := false
	moveToNewTopic := false

	if chatGPTResponse.NextQuestion == "finished" {
		isFinished = true
	}

	if chatGPTResponse.Topic != PredefinedTopics[conversation.CurrentTopic].Name {
		moveToNewTopic = true
	}

	return moveToNewTopic, chatGPTResponse.MoveToNewSubtopic, isFinished
}
