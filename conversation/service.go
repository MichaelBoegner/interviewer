package conversation

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
)

func CheckForConversation(repo ConversationRepo, interviewID int) bool {
	return repo.CheckForConversation(interviewID)
}

func CreateConversation(
	repo ConversationRepo,
	openAI chatgpt.AIClient,
	interviewID int,
	prompt,
	firstQuestion,
	subtopic,
	message string) (*Conversation, error) {
	now := time.Now().UTC()

	conversation := &Conversation{
		InterviewID:           interviewID,
		Topics:                PredefinedTopics,
		CurrentTopic:          1,
		CurrentSubtopic:       subtopic,
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
	_, err = repo.CreateQuestion(conversation, firstQuestion)
	if err != nil {
		log.Printf("CreateQuestion failing")
		return nil, err
	}

	topic := conversation.Topics[conversation.CurrentTopic]
	topic.ConversationID = conversationID
	messages := []Message{
		newMessage(conversationID, conversation.CurrentTopic, conversation.CurrentQuestionNumber, System, prompt),
		newMessage(conversationID, conversation.CurrentTopic, conversation.CurrentQuestionNumber, Interviewer, firstQuestion),
		newMessage(conversationID, conversation.CurrentTopic, conversation.CurrentQuestionNumber, User, message),
	}
	question := newQuestion(conversationID, conversation.CurrentTopic, conversation.CurrentQuestionNumber, messages)
	topic.Questions = make(map[int]*Question)
	topic.Questions[conversation.CurrentQuestionNumber] = question
	conversation.Topics[conversation.CurrentTopic] = topic

	err = repo.CreateMessages(conversation, question.Messages)
	if err != nil {
		log.Printf("repo.CreateMessages failing")
		return nil, err
	}

	conversationHistory, err := getConversationHistory(conversation)
	if err != nil {
		return nil, err
	}
	chatGPTResponse, err := openAI.GetChatGPTResponseConversation(conversationHistory)
	if err != nil {
		log.Printf("getNextQuestion failing")
		return nil, err
	}
	chatGPTResponseString, err := ChatGPTResponseToString(chatGPTResponse)
	if err != nil {
		log.Printf("Marshalled response err: %v", err)
		return nil, err
	}

	conversation.CurrentQuestionNumber++
	_, err = repo.UpdateConversationCurrents(conversation.ID, conversation.CurrentTopic, conversation.CurrentQuestionNumber, chatGPTResponse.NextSubtopic)
	if err != nil {
		log.Printf("UpdateConversationTopic error: %v", err)
		return nil, err
	}

	messages = []Message{}
	conversation.Topics[conversation.CurrentTopic].Questions[conversation.CurrentQuestionNumber] = newQuestion(conversationID, conversation.CurrentTopic, conversation.CurrentQuestionNumber, messages)
	messageNextQuestion := newMessage(conversationID, conversation.CurrentTopic, conversation.CurrentQuestionNumber, Interviewer, chatGPTResponseString)

	topic = conversation.Topics[conversation.CurrentTopic]
	topic.Questions[conversation.CurrentQuestionNumber].Messages = append(topic.Questions[conversation.CurrentQuestionNumber].Messages, messageNextQuestion)
	conversation.Topics[conversation.CurrentTopic] = topic

	_, err = repo.AddQuestion(conversation.Topics[conversation.CurrentTopic].Questions[conversation.CurrentQuestionNumber])
	if err != nil {
		log.Printf("AddQuestion in AppendConversation err: %v", err)
		return nil, err
	}
	_, err = repo.AddMessage(conversationID, conversation.CurrentTopic, conversation.CurrentQuestionNumber, messageNextQuestion)
	if err != nil {
		log.Printf("AddMessage in AppendConversation err: %v", err)
		return nil, err
	}

	return conversation, nil
}

func AppendConversation(
	repo ConversationRepo,
	openAI chatgpt.AIClient,
	conversation *Conversation,
	message, prompt string) (*Conversation, error) {

	conversationID := conversation.ID
	topicID := conversation.CurrentTopic
	questionNumber := conversation.CurrentQuestionNumber

	if conversation.ID != conversationID {
		return nil, errors.New("conversation_id doesn't match with current interview")
	}

	messageUser := newMessage(conversationID, topicID, questionNumber, User, message)
	_, err := repo.AddMessage(conversationID, topicID, questionNumber, messageUser)
	if err != nil {
		return nil, err
	}
	conversation.Topics[topicID].Questions[questionNumber].Messages = append(conversation.Topics[topicID].Questions[questionNumber].Messages, messageUser)

	conversationHistory, err := getConversationHistory(conversation)
	if err != nil {
		return nil, err
	}
	chatGPTResponse, err := openAI.GetChatGPTResponseConversation(conversationHistory)
	if err != nil {
		log.Printf("getNextQuestion err: %v", err)
		return nil, err
	}
	chatGPTResponseString, err := ChatGPTResponseToString(chatGPTResponse)
	if err != nil {
		log.Printf("Marshalled response err: %v", err)
		return nil, err
	}

	moveToNewTopic, incrementQuestion, isFinished, err := checkConversationState(chatGPTResponse, conversation)
	if err != nil {
		log.Printf("checkConversationState err: %v", err)
		return nil, err
	}

	if isFinished {
		conversation.CurrentTopic = 0
		conversation.CurrentSubtopic = "Finished"
		conversation.CurrentQuestionNumber = 0

		_, err := repo.UpdateConversationCurrents(conversationID, conversation.CurrentTopic, 0, conversation.CurrentSubtopic)
		if err != nil {
			log.Printf("UpdateConversationTopic error: %v", err)
			return nil, err
		}

		messageFinal := newMessage(conversationID, conversation.CurrentTopic, questionNumber, Interviewer, chatGPTResponseString)
		_, err = repo.AddMessage(conversationID, topicID, questionNumber, messageFinal)
		if err != nil {
			return nil, err
		}

		conversation.Topics[topicID].Questions[questionNumber].Messages = append(conversation.Topics[topicID].Questions[questionNumber].Messages, messageFinal)

		return conversation, nil
	}

	if moveToNewTopic {
		nextTopicID := topicID + 1
		resetQuestionNumber := 1
		conversation.CurrentTopic = nextTopicID
		conversation.CurrentSubtopic = chatGPTResponse.NextSubtopic
		conversation.CurrentQuestionNumber = 1

		_, err := repo.UpdateConversationCurrents(conversationID, conversation.CurrentTopic, conversation.CurrentQuestionNumber, conversation.CurrentSubtopic)
		if err != nil {
			log.Printf("UpdateConversationTopic error: %v", err)
			return nil, err
		}

		topic := conversation.Topics[nextTopicID]
		topic.ConversationID = conversationID

		messageFirstQuestion := newMessage(conversationID, conversation.CurrentTopic, resetQuestionNumber, Interviewer, chatGPTResponseString)
		messages := []Message{
			messageFirstQuestion,
		}
		question := newQuestion(conversationID, topicID, questionNumber, messages)

		topic.Questions = make(map[int]*Question)
		topic.Questions[resetQuestionNumber] = question

		conversation.Topics[nextTopicID] = topic

		_, err = repo.AddQuestion(question)
		if err != nil {
			log.Printf("AddQuestion in AppendConversation err: %v", err)
		}
		_, err = repo.AddMessage(conversationID, nextTopicID, resetQuestionNumber, messageFirstQuestion)
		if err != nil {
			return nil, err
		}

		return conversation, nil
	}

	if incrementQuestion {
		conversation.CurrentQuestionNumber++
		questionNumber++
		_, err := repo.UpdateConversationCurrents(conversationID, topicID, questionNumber, chatGPTResponse.NextSubtopic)
		if err != nil {
			log.Printf("UpdateConversationTopic error: %v", err)
			return nil, err
		}
		messages := []Message{}
		conversation.Topics[topicID].Questions[questionNumber] = newQuestion(conversationID, topicID, questionNumber, messages)
	}

	messageNextQuestion := newMessage(conversationID, topicID, questionNumber, Interviewer, chatGPTResponseString)
	conversation.Topics[topicID].Questions[questionNumber].Messages = append(conversation.Topics[topicID].Questions[questionNumber].Messages, messageNextQuestion)

	_, err = repo.AddQuestion(conversation.Topics[topicID].Questions[questionNumber])
	if err != nil {
		log.Printf("AddQuestion in AppendConversation err: %v", err)
		return nil, err
	}
	_, err = repo.AddMessage(conversationID, topicID, questionNumber, messageNextQuestion)
	if err != nil {
		log.Printf("AddMessage in AppendConversation err: %v", err)
		return nil, err
	}

	return conversation, nil
}

func GetConversation(repo ConversationRepo, conversationID int) (*Conversation, error) {
	// Get conversation from Conversations table and apply to Conversation struct
	conversation, err := repo.GetConversation(conversationID)
	if err != nil {
		return nil, err
	}

	// Add Topic structs to conversation
	conversation.Topics = PredefinedTopics

	// Get questions from Questions table to apply to conversation.topics
	questionsReturned, err := repo.GetQuestions(conversation)
	if err != nil {
		return nil, err
	}

	// Apply returned questions to respective Topic structs
	for topicID := 1; topicID <= conversation.CurrentTopic; topicID++ {
		topic := conversation.Topics[topicID]
		topic.ConversationID = conversation.ID
		topic.Questions = make(map[int]*Question)

		for _, question := range questionsReturned {
			if question.TopicID != topicID {
				continue
			}

			topic.Questions[question.QuestionNumber] = question

			messagesReturned, err := repo.GetMessages(conversation.ID, topicID, question.QuestionNumber)
			if err != nil {
				log.Printf("repo.GetMessages failed: %v\n", err)
				return nil, err
			}

			question.Messages = append(question.Messages, messagesReturned...)

			conversation.Topics[topicID] = topic
			conversation.Topics[topicID].Questions[question.QuestionNumber] = question
		}
	}

	return conversation, nil
}

func getConversationHistory(conversation *Conversation) ([]map[string]string, error) {
	chatGPTConversationArray := make([]map[string]string, 0)

	var arrayOfTopics []string
	var currentTopic string

	currentTopic = PredefinedTopics[conversation.CurrentTopic].Name

	for topic := 1; topic < conversation.CurrentTopic; topic++ {
		arrayOfTopics = append(arrayOfTopics, PredefinedTopics[topic].Name)
	}

	systemPrompt := map[string]string{
		"role": "system",
		"content": fmt.Sprintf("You are conducting a structured backend development interview. "+
			"The interview follows **six topics in this order**:\n\n"+
			"1. **Introduction**\n"+
			"2. **Coding**\n"+
			"3. **System Design**\n"+
			"4. **Databases**\n"+
			"5. **Behavioral**\n"+
			"6. **General Backend Knowledge**\n\n"+
			"You have already covered the following topics: %s.\n"+
			"You are currently on the topic: %s. \n\n"+
			"**Rules:**\n"+
			"- Ask **exactly 2 questions per topic** before moving to the next.\n"+
			"- Do **not** skip or reorder topics.\n"+
			"- You only have access to the current topic’s conversation history. Infer progression logically.\n"+
			"- Format responses as **valid JSON only** (no explanations or extra text).\n\n"+
			"**If candidate says 'I don't know':**\n"+
			"- Assign **score: 1** and provide minimal feedback.\n"+
			"- Move to the next question.\n\n"+
			"**JSON Response Format:**\n"+
			"{\n"+
			"    \"topic\": \"current topic\",\n"+
			"    \"subtopic\": \"current subtopic\",\n"+
			"    \"question\": \"previous question\",\n"+
			"    \"score\": the score (1-10) you think the previous answer deserves, default to 0 if you don't have a score,\n"+
			"    \"feedback\": \"brief feedback\",\n"+
			"    \"next_question\": \"next question\",\n"+
			"    \"next_topic\": \"next topic\",\n"+
			"    \"next_subtopic\": \"next subtopic\"\n"+
			"}", arrayOfTopics, currentTopic),
	}

	chatGPTConversationArray = append(chatGPTConversationArray, systemPrompt)

	topic := conversation.Topics[conversation.CurrentTopic]

	if len(topic.Questions) == 0 {
		return nil, errors.New("no questions found in conversation")
	}

	// Iterate through all questions within the current topic
	for _, question := range topic.Questions {
		for i, message := range question.Messages {
			// Skip the system prompt message (only if it's the very first message in the first question)
			if conversation.CurrentTopic == 1 && conversation.CurrentQuestionNumber == 1 && i == 0 {
				continue
			}

			role := "user"
			if message.Author == "interviewer" {
				role = "assistant"
			}

			chatGPTConversationArray = append(chatGPTConversationArray, map[string]string{
				"role":    role,
				"content": message.Content,
			})
		}
	}

	return chatGPTConversationArray, nil
}

func newMessage(conversationID, topicID, currentQuestionNumber int, author Author, content string) Message {
	message := Message{
		ConversationID: conversationID,
		QuestionNumber: currentQuestionNumber,
		TopicID:        topicID,
		Author:         author,
		Content:        content,
		CreatedAt:      time.Now().UTC(),
	}

	return message
}

func newQuestion(conversationID, topicID, currentQuestionNumber int, messages []Message) *Question {
	return &Question{
		ConversationID: conversationID,
		TopicID:        topicID,
		QuestionNumber: currentQuestionNumber,
		Messages:       messages,
		CreatedAt:      time.Now().UTC(),
	}
}

func ChatGPTResponseToString(chatGPTResponse *chatgpt.ChatGPTResponse) (string, error) {
	chatGPTResponseString, err := json.Marshal(chatGPTResponse)
	if err != nil {
		log.Printf("chatGPTResponseToString failed: %v", err)
		return "", err
	}

	return string(chatGPTResponseString), nil
}

func checkConversationState(chatGPTResponse *chatgpt.ChatGPTResponse, conversation *Conversation) (bool, bool, bool, error) {
	isFinished := false
	moveToNewTopic := false
	incrementQuestion := false

	if chatGPTResponse.NextTopic != PredefinedTopics[conversation.CurrentTopic].Name {
		moveToNewTopic = true
	}

	if chatGPTResponse.NextSubtopic != conversation.CurrentSubtopic {
		incrementQuestion = true
	}

	if chatGPTResponse.Topic == "General Backend Knowledge" {
		isFinished = true
	}

	return moveToNewTopic, incrementQuestion, isFinished, nil
}
