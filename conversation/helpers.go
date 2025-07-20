package conversation

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/interview"
)

func GetChatGPTResponses(conversation *Conversation, openAI chatgpt.AIClient, interviewRepo interview.InterviewRepo, conversationContext []string) (*chatgpt.ChatGPTResponse, string, error) {
	conversationHistory, err := GetConversationHistory(conversation, interviewRepo, conversationContext)
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

func GetConversationHistory(conversation *Conversation, interviewRepo interview.InterviewRepo, conversationContext []string) ([]map[string]string, error) {
	var arrayOfTopics []string
	var currentTopic string
	chatGPTConversationArray := make([]map[string]string, 0)
	predefinedTopics := ClonePredefinedTopics()

	currentTopic = predefinedTopics[conversation.CurrentTopic].Name
	for topic := 1; topic < conversation.CurrentTopic; topic++ {
		arrayOfTopics = append(arrayOfTopics, predefinedTopics[topic].Name)
	}

	interview, err := interviewRepo.GetInterview(conversation.InterviewID)
	if err != nil {
		log.Printf("jdsummary.JDCache.Get() did not return anything")
		return nil, err
	}

	systemPrompt := map[string]string{
		"role":    "system",
		"content": chatgpt.BuildPrompt(arrayOfTopics, currentTopic, conversation.CurrentQuestionNumber, interview.JDSummary),
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
	lastQuestionNumber := questionNumbersSorted[len(questionNumbersSorted)-1]
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

			content := message.Content
			isFinalInjectionTarget := questionNumber == lastQuestionNumber &&
				message.Author == "user"
				// DEBUG
			fmt.Printf("isFinalInjectionTarget: %v\n", isFinalInjectionTarget)
			fmt.Printf("conversationContext: %v\n", conversationContext)
			if isFinalInjectionTarget && len(conversationContext) > 0 {
				formattedContext := strings.Join(conversationContext, "\n")
				content = fmt.Sprintf("Relevant prior user context:\n%s\n\n--- BEGIN USER'S ACTUAL RESPONSE ---\n%s", formattedContext, content)
			}

			chatGPTConversationArray = append(chatGPTConversationArray, map[string]string{
				"role":    role,
				"content": content,
			})
		}
	}

	fmt.Println("------ DEBUG: Formatted Conversation History ------")
	for i, msg := range chatGPTConversationArray {
		fmt.Printf("\n--- Message %d ---\n", i+1)
		fmt.Printf("Role   : %s\n", msg["role"])
		fmt.Printf("Content:\n%s\n", msg["content"])
	}
	fmt.Println("------ END DEBUG ------")

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
			Questions: make(map[int]*Question),
		}
	}
	return topics
}
