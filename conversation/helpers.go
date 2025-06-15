package conversation

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
)

func GetChatGPTResponses(conversation *Conversation, openAI chatgpt.AIClient) (*chatgpt.ChatGPTResponse, string, error) {
	conversationHistory, err := GetConversationHistory(conversation)
	if err != nil {
		log.Printf("GetConversationHistory failed: %v", err)
		return nil, "", err
	}
	chatGPTResponse, err := openAI.GetChatGPTResponseConversation(conversationHistory)
	if err != nil {
		log.Printf("getNextQuestion failed: %v", err)
		return nil, "", err
	}
	chatGPTResponseString, err := ChatGPTResponseToString(chatGPTResponse)
	if err != nil {
		log.Printf("Marshalled response failed: %v", err)
		return nil, "", err
	}

	return chatGPTResponse, chatGPTResponseString, nil
}

func GetConversationHistory(conversation *Conversation) ([]map[string]string, error) {
	chatGPTConversationArray := make([]map[string]string, 0)

	var arrayOfTopics []string
	var currentTopic string

	predefinedTopics := ClonePredefinedTopics()

	currentTopic = predefinedTopics[conversation.CurrentTopic].Name

	for topic := 1; topic < conversation.CurrentTopic; topic++ {
		arrayOfTopics = append(arrayOfTopics, predefinedTopics[topic].Name)
	}

	systemPrompt := map[string]string{
		"role": "system",
		"content": fmt.Sprintf("You are conducting a structured coding-language-agnostic, backend development interview. "+
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
			"    \"score\": the score (1-10) you think the previous answer deserves. Treat a score of 7 as the minimum passing threshold. Only give 8–10 for answers that are complete, technically sound, and reflect senior-level expertise. Use scores 1–6 freely to reflect any gaps, vagueness, or missed edge cases. Default to 0 if no score is possible,\n"+
			"    \"feedback\": \"keep feedback brief\",\n"+
			"    \"next_topic\": \"next topic based strictly on the above topic list\",\n"+
			"    \"next_subtopic\": \"next subtopic\"\n"+
			"    \"next_question\": \"next question\",\n"+
			"}", arrayOfTopics, currentTopic),
	}

	chatGPTConversationArray = append(chatGPTConversationArray, systemPrompt)

	topic := conversation.Topics[conversation.CurrentTopic]

	if len(topic.Questions) == 0 {
		return nil, errors.New("no questions found in conversation")
	}

	questionNumbersSorted := make([]int, 0, len(topic.Questions))
	for questionNumber := range topic.Questions {
		questionNumbersSorted = append(questionNumbersSorted, questionNumber)
	}
	sort.Ints(questionNumbersSorted)

	for _, questionNumber := range questionNumbersSorted {
		question := topic.Questions[questionNumber]
		for i, message := range question.Messages {
			if conversation.CurrentTopic == 1 && conversation.CurrentQuestionNumber == 1 && i == 0 {
				continue
			}
			if message.Author == "system" {
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

func NewMessage(conversationID, topicID, currentQuestionNumber int, author Author, content string) Message {
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

func NewQuestion(conversationID, topicID, currentQuestionNumber int, prompt string, messages []Message) *Question {
	return &Question{
		ConversationID: conversationID,
		TopicID:        topicID,
		QuestionNumber: currentQuestionNumber,
		Prompt:         prompt,
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

func CheckConversationState(chatGPTResponse *chatgpt.ChatGPTResponse, conversation *Conversation) (bool, bool, bool, error) {
	topic := conversation.Topics[conversation.CurrentTopic]
	questionCount := len(topic.Questions)

	isFinished := chatGPTResponse.Topic == "General Backend Knowledge" && questionCount == 2
	switch {
	case questionCount >= 2:
		return true, false, isFinished, nil
	case questionCount == 1:
		return false, true, isFinished, nil
	default:
		return false, false, isFinished, nil
	}
}

func ClonePredefinedTopics() map[int]Topic {
	topics := make(map[int]Topic)
	for id, topic := range PredefinedTopics {
		topics[id] = Topic{
			ID:        topic.ID,
			Name:      topic.Name,
			Questions: make(map[int]*Question), // fresh for each conversation
		}
	}
	return topics
}
