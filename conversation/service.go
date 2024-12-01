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

	"github.com/michaelboegner/interviewer/interview"
)

func CheckForConversation(repo ConversationRepo, interviewID int) bool {
	return repo.CheckForConversation(interviewID)
}

func CreateConversation(
	repo ConversationRepo,
	interviewID int,
	firstQuestion string,
	questionContext *interview.QuestionContext,
	messageUserResponse *Message) (*Conversation, error) {

	if messageUserResponse == nil {
		log.Printf("messageUserResponse is nil")
		return nil, errors.New("messageUserResponse cannot be nil")
	}

	now := time.Now()
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

	questionContextString, err := questionContextToString(questionContext)
	if err != nil {
		log.Printf("questionContextToString failed: %v", err)
		return nil, err
	}

	messageFirst := newMessage(1, questionID, Interviewer, questionContextString)

	messageUserResponse.ID = 2
	messageUserResponse.QuestionID = questionID
	messageUserResponse.CreatedAt = time.Now()

	question := &Question{
		ID:             questionID,
		QuestionNumber: 1,
		ConversationID: conversationID,
		TopicID:        1,
		Prompt:         firstQuestion,
		Messages: []Message{
			*messageFirst,
			*messageUserResponse,
		},
		CreatedAt: time.Now(),
	}

	topic.Questions = make(map[int]*Question)
	topic.Questions[1] = question

	conversation.Topics[1] = topic

	nextQuestion, err := getNextQuestion(conversation, 1, 1)
	if err != nil {
		log.Printf("getNextQuestion failing")
		return nil, err
	}

	topic = conversation.Topics[1]

	topic.Questions[1].Messages = append(topic.Questions[1].Messages, *newMessage(3, questionID, User, nextQuestion))

	conversation.Topics[1] = topic

	err = repo.CreateMessages(conversation, question.Messages)
	if err != nil {
		log.Printf("repo.CreateMessages failing")
		return nil, err
	}

	return conversation, nil
}

func AppendConversation(repo ConversationRepo, conversation *Conversation, message *Message, conversationID, topicID, questionID, questionNumber int) (*Conversation, error) {
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

	nextQuestion, err := getNextQuestion(conversation, topicID, questionNumber)
	if err != nil {
		log.Printf("getNextQuestion failing")
		return nil, err
	}

	messageNextQuestionID := messageID + 1
	messageNextQuestion := newMessage(messageNextQuestionID, questionID, Interviewer, nextQuestion)

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

func getNextQuestion(conversation *Conversation, topicID, questionNumber int) (string, error) {
	ctx := context.Background()
	apiKey := os.Getenv("OPENAI_API_KEY")

	prompt := "You are conducting a technical interview for a backend development position. " +
		"The interview is divided into six main topics:\n\n" +
		"1. **Introduction**\n" +
		"2. **Coding**\n" +
		"3. **System Design**\n" +
		"4. **Databases and Data Management**\n" +
		"5. **Behavioral**\n" +
		"6. **General Backend Knowledge**\n\n" +
		"Each main topic may contain multiple subtopics. For each subtopic:\n" +
		"- Ask as many subtopic-specific questions as needed until the candidate has " +
		"sufficiently or insufficiently proven their understanding of the subtopic.\n" +
		"- Provide a score (1-10) for their performance on each subtopic question, with feedback " +
		"explaining the score.\n" +
		"- When the subtopic is sufficiently assessed, decide if the current topic needs " +
		"further exploration or if it's time to move to the next topic.\n\n" +
		"After each question:\n" +
		"1. Wait for the candidate's response before evaluating their answer.\n" +
		"2. If the candidate has not yet responded, return blank values for \"score\", \"feedback\", and \"next_question\".\n" +
		"3. Evaluate the candidate's response and decide if additional questions about the " +
		"current subtopic are necessary.\n" +
		"4. If transitioning to a new subtopic or topic, include a flag (move_to_new_subtopic) " +
		"and/or (move_to_new_topic) in the response. Otherwise, set these flags to false.\n" +
		"5. Return your response in the following JSON format:\n\n" +
		"{\n" +
		"    \"topic\": \"System Design\",\n" +
		"    \"subtopic\": \"Scalability\",\n" +
		"    \"question\": \"How would you design a system to handle a high number of concurrent users?\",\n" +
		"    \"score\": null,\n" +
		"    \"feedback\": \"\",\n" +
		"    \"next_question\": \"\",\n" +
		"    \"move_to_new_subtopic\": false,\n" +
		"    \"move_to_new_topic\": false\n" +
		"}"

	systemMessage := map[string]string{
		"role":    "system",
		"content": prompt,
	}

	conversationHistory, err := getConversationHistory(conversation, topicID, questionNumber)
	if err != nil {
		return "", err
	}

	completeConversation := append([]map[string]string{systemMessage}, conversationHistory...)

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       "gpt-4",
		"messages":    completeConversation,
		"max_tokens":  150,
		"temperature": 0.7,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("NewRequestWithContext failing")
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	responseChan := make(chan string)
	errorChan := make(chan error)

	log.Printf("Request body: %s", requestBody)

	go func() {
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errorChan <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("API call failed with status code: %d, response: %s", resp.StatusCode, body)
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

func questionContextToString(questionContext *interview.QuestionContext) (string, error) {
	questionContextString, err := json.Marshal(questionContext)
	if err != nil {
		log.Printf("questionContextToString failed: %v", err)
		return "", err
	}

	return string(questionContextString), nil
}
