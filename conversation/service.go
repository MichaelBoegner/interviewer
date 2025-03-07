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
	interviewID,
	questionNumber int,
	prompt,
	firstQuestion,
	subtopic string,
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
		CurrentSubtopic:       subtopic,
		CurrentQuestionNumber: questionNumber,
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

	// Check that conversation ID matches with Interview ID
	if conversation.ID != conversationID {
		return nil, errors.New("conversation_id doesn't match with current interview")
	}

	// Add response message from user to Messages table
	_, err := repo.AddMessage(conversationID, conversation.CurrentTopic, questionNumber, message)
	if err != nil {
		return nil, err
	}

	// Add response message to Messages struct
	messageUser := newMessage(conversationID, questionNumber, message.Author, message.Content)

	messages := conversation.Topics[topicID].Questions[questionNumber].Messages
	messages = append(messages, *messageUser)
	conversation.Topics[topicID].Questions[questionNumber].Messages = messages

	// Call ChatGPT for next question and convert to string and store. String conversion is need for when sending convo history back to ChatGPT.
	chatGPTResponse, err := getNextQuestion(conversation)
	if err != nil {
		log.Printf("getNextQuestion err: %v", err)
		return nil, err
	}

	chatGPTResponseString, err := ChatGPTResponseToString(chatGPTResponse)
	if err != nil {
		log.Printf("Marshalled response err: %v", err)
		return nil, err
	}
	//DEBUG PRINT CONVERSATION TOPIC AND SUBTOPIC STATES
	fmt.Printf("conversation.CurrentTopic state being checked: %v\n", conversation.CurrentTopic)
	fmt.Printf("conversation.CurrentSubtopic state being checked: %v\n", conversation.CurrentSubtopic)
	fmt.Printf("chatGPTResponse.Topic state being checked: %v\n", chatGPTResponse.NextTopic)
	fmt.Printf("chatGPTResponse.Subtopic state being checked: %v\n", chatGPTResponse.NextSubtopic)

	// Check the current states of the Conversation
	moveToNewTopic, incrementQuestion, isFinished, err := checkConversationState(chatGPTResponse, conversation, repo)
	if err != nil {
		log.Printf("checkConversationState err: %v", err)
		return nil, err
	}

	// if isFinished, then raise flags to close conversation and begin closing sequence
	if isFinished {
		conversation.CurrentTopic = 0
		conversation.CurrentSubtopic = "Finished"
		conversation.CurrentQuestionNumber = 0

		_, err := repo.UpdateConversationCurrents(conversationID, conversation.CurrentTopic, 0, conversation.CurrentSubtopic)
		if err != nil {
			log.Printf("UpdateConversationTopic error: %v", err)
			return nil, err
		}

		return conversation, nil
	}

	// if moveToNewTopic, increment topicID and reset questionNumber
	if moveToNewTopic {
		//DEBUG PRINT MOVETONEWTOPIC AND CONVERSATION TOPIC AND CHATGPT NEXT TOPIC
		fmt.Printf("\n\n\nmoveToNewTopic: %v\n", moveToNewTopic)
		fmt.Printf("conversation.CurrentTopic: %v\n", conversation.CurrentTopic)
		fmt.Printf("chatGPTResponse.NextTopic: %v\n", chatGPTResponse.NextTopic)
		fmt.Printf("conversation.CurrentQuestionNumber: %v\n", conversation.CurrentQuestionNumber)
		fmt.Printf("questionNumber: %v\n\n\n", questionNumber)

		nextTopicID := topicID + 1
		resetQuestionNumber := 1
		_, err := repo.UpdateConversationCurrents(conversationID, nextTopicID, resetQuestionNumber, chatGPTResponse.Subtopic)
		if err != nil {
			log.Printf("UpdateConversationTopic error: %v", err)
			return nil, err
		}

		topic := conversation.Topics[nextTopicID]
		topic.ConversationID = conversationID

		messageFirstQuestion := newMessage(conversationID, resetQuestionNumber, Interviewer, chatGPTResponseString)

		question := &Question{
			ConversationID: conversationID,
			TopicID:        nextTopicID,
			QuestionNumber: resetQuestionNumber,
			Prompt:         chatGPTResponse.NextQuestion,
			Messages: []Message{
				*messageFirstQuestion,
			},
			CreatedAt: time.Now(),
		}

		topic.Questions = make(map[int]*Question)
		topic.Questions[resetQuestionNumber] = question

		conversation.Topics[nextTopicID] = topic

		_, err = repo.AddQuestion(question)
		if err != nil {
			log.Printf("AddQuestion in AppendConversation err: %v", err)
		}

		_, err = repo.AddMessage(conversationID, nextTopicID, questionNumber, messageFirstQuestion)
		if err != nil {
			return nil, err
		}

		return conversation, nil
	}

	// If not new Topic, then continue building under current topic and return conversation
	if incrementQuestion {
		conversation.CurrentQuestionNumber++
		_, err := repo.UpdateConversationCurrents(conversation.ID, conversation.CurrentTopic, conversation.CurrentQuestionNumber, chatGPTResponse.NextSubtopic)
		if err != nil {
			log.Printf("UpdateConversationTopic error: %v", err)
			return nil, err
		}

		questionNumber += 1

		if conversation.Topics[topicID].Questions[questionNumber] == nil {
			conversation.Topics[topicID].Questions[questionNumber] = &Question{
				ConversationID: conversation.ID,
				TopicID:        topicID,
				QuestionNumber: questionNumber,
				Messages:       []Message{},
				CreatedAt:      time.Now(),
			}
		}
	}
	//DEBUG CHECK INCREMENTQUESTION, CURRENTQUESTIONUMBER, AND QUESTIONNUMBER
	fmt.Printf("\n\n\nincrementQuestion: %v\n", incrementQuestion)
	fmt.Printf("conversation.CurrentQuestionNumber: %v\n", conversation.CurrentQuestionNumber)
	fmt.Printf("questionNumber: %v\n\n\n", questionNumber)

	messageNextQuestion := newMessage(conversationID, questionNumber, Interviewer, chatGPTResponseString)

	messages = conversation.Topics[topicID].Questions[questionNumber].Messages
	messages = append(messages, *messageNextQuestion)
	conversation.Topics[topicID].Questions[questionNumber].Messages = messages

	_, err = repo.AddQuestion(conversation.Topics[topicID].Questions[questionNumber])
	if err != nil {
		log.Printf("AddQuestion in AppendConversation err: %v", err)
		return nil, err
	}

	_, err = repo.AddMessage(conversationID, conversation.CurrentTopic, questionNumber, messageNextQuestion)
	if err != nil {
		log.Printf("AddMessage in AppendConversation err: %v", err)
		return nil, err
	}

	return conversation, nil
}

func GetConversation(repo ConversationRepo, interviewID int) (*Conversation, error) {
	// Get conversation from Conversations table and apply to Conversation struct
	conversation, err := repo.GetConversation(interviewID)
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

	//DEBUG PRINT QUESTIONS RETURNED
	for i, q := range questionsReturned {
		fmt.Printf("\nquestionsReturned[%d]: \n%+v\n", i, *q)
	}

	// Apply returned questions to respective Topic structs
	for topicID := 1; topicID <= conversation.CurrentTopic; topicID++ {
		fmt.Printf("topicID: %v\n", topicID)

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

	var arrayOfTopics []string
	var currentTopic string

	currentTopic = PredefinedTopics[conversation.CurrentTopic].Name

	for topic := 1; topic < conversation.CurrentTopic; topic++ {
		arrayOfTopics = append(arrayOfTopics, PredefinedTopics[topic].Name)
	}

	systemPrompt := map[string]string{
		"role": "system",
		"content": fmt.Sprintf("You are conducting a technical interview for a backend development position. "+
			"There are **six topics**, which **must be followed in this order**:\n\n"+
			"1. **Introduction**\n"+
			"2. **Coding**\n"+
			"3. **System Design**\n"+
			"4. **Databases and Data Management**\n"+
			"5. **Behavioral**\n"+
			"6. **General Backend Knowledge**\n\n"+
			"The interview must follow the topics in the exact order specified above (1 to 6).\n\n"+
			"You have already covered the following topics: %s.\n"+
			"You are currently on the topic: %s. \n\n"+
			"Do not skip any topic, even if the candidate's performance suggests otherwise. "+
			"Ensure the next question is always relevant to the current topic or subtopic until it is fully assessed, "+
			"then proceed to the next topic in the order. Do not ask more than 2 questions per topic.\n\n"+
			"### **CONVERSATION HISTORY LIMITATIONS**\n"+
			"You are only being provided with the **entire conversation history for the current topic**. "+
			"You do **NOT** have access to previous topics. "+
			"Based on this, infer the next **subtopic and question** while strictly maintaining topic order.\n\n"+
			"### **STRICT TOPIC ADHERENCE**\n"+
			"You must never jump ahead, skip, or alter the sequence of topics. "+
			"If the candidate completes the current topic, move **only to the next topic in order**. "+
			"If you are uncertain of the context due to missing history, assume normal topic progression and "+
			"generate the next most logical question.\n\n"+
			"### **STRICT JSON-ONLY RESPONSE ENFORCEMENT**\n"+
			"1. **You must ALWAYS return a valid JSON object.** Never respond conversationally.\n"+
			"2. **DO NOT provide explanations, encouragement, or assistant-style messages.**\n"+
			"3. **DO NOT generate additional text outside of the JSON format.** Any response outside of JSON format is strictly forbidden.\n\n"+
			"### **Handling 'I Don't Know' Responses**\n"+
			"1. **If the candidate responds with 'I don't know' or an equivalent phrase:**\n"+
			"   - Assign the lowest appropriate score (e.g., 1) for the question.\n"+
			"   - Provide structured feedback stating that the candidate did not provide an answer.\n"+
			"   - Immediately proceed to the next relevant question while maintaining strict topic order.\n"+
			"   - **DO NOT generate any conversational, helpful, or assistant-like responses.**\n\n"+
			"### **Topic Transition Rule**\n"+
			"**If transitioning to a new topic, remind yourself that this is still part of the structured interview. "+
			"The JSON format must remain consistent across all topics. DO NOT break out of JSON at any point.**\n\n"+
			"Your response must **ALWAYS** follow this format:\n\n"+
			"{\n"+
			"    \"topic\": \"the current topic\",\n"+
			"    \"subtopic\": \"the current subtopic\",\n"+
			"    \"question\": \"the previous question\",\n"+
			"    \"score\": the score (1-10) you think the previous answer deserves,\n"+
			"    \"feedback\": \"your feedback about the quality of the previous answer\",\n"+
			"    \"next_question\": \"the next question\",\n"+
			"    \"next_topic\": \"the topic of the next question\",\n"+
			"    \"next_subtopic\": \"the subtopic of the next question\",\n"+
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

	// Debugging output
	prettyJSON, _ := json.MarshalIndent(chatGPTConversationArray, "", "  ")
	fmt.Println("THIS IS WHAT YOU'RE SENDING TO CHATGPT TO GET THE NEXT QUESTION:")
	fmt.Println(string(prettyJSON))

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

func checkConversationState(chatGPTResponse *models.ChatGPTResponse, conversation *Conversation, repo ConversationRepo) (bool, bool, bool, error) {
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
