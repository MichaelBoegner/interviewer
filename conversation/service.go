package conversation

import (
	"errors"
	"log"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/interview"
)

func CheckForConversation(repo ConversationRepo, interviewID int) (bool, error) {
	return repo.CheckForConversation(interviewID)
}

func CreateEmptyConversation(repo ConversationRepo, interviewID int, subTopic string) (int, error) {
	conversation := &Conversation{
		Topics:                ClonePredefinedTopics(),
		CurrentTopic:          1,
		CurrentSubtopic:       subTopic,
		CurrentQuestionNumber: 1,
	}

	conversationID, err := repo.CreateConversation(interviewID, conversation)
	if err != nil {
		log.Printf("CreateConversation failed: %v", err)
		return 0, err
	}

	return conversationID, nil
}

func CreateConversation(
	repo ConversationRepo,
	interviewRepo interview.InterviewRepo,
	openAI chatgpt.AIClient,
	conversation *Conversation,
	interviewID int,
	prompt,
	firstQuestion,
	subtopic,
	message string) (*Conversation, error) {

	conversationID := conversation.ID
	topicID := conversation.CurrentTopic
	questionNumber := conversation.CurrentQuestionNumber

	_, err := repo.CreateQuestion(conversation, firstQuestion)
	if err != nil {
		log.Printf("CreateQuestion failed: %v", err)
		return nil, err
	}

	topic := conversation.Topics[conversation.CurrentTopic]
	topic.ConversationID = conversationID
	messages := []Message{
		NewMessage(conversationID, topicID, questionNumber, System, prompt),
		NewMessage(conversationID, topicID, questionNumber, Interviewer, firstQuestion),
		NewMessage(conversationID, topicID, questionNumber, User, message),
	}
	topic.Questions = make(map[int]*Question)
	topic.Questions[questionNumber] = NewQuestion(conversationID, topicID, questionNumber, firstQuestion, messages)
	conversation.Topics[topicID] = topic

	err = repo.CreateMessages(conversation, messages)
	if err != nil {
		log.Printf("repo.CreateMessages failed: %v", err)
		return nil, err
	}

	chatGPTResponse, chatGPTResponseString, err := GetChatGPTResponses(conversation, openAI, interviewRepo, firstQuestion, message)
	if err != nil {
		log.Printf("getChatGPTResponses failed: %v", err)
		return nil, err
	}

	err = interviewRepo.UpdateScore(interviewID, chatGPTResponse.Score)
	if err != nil {
		log.Printf("interviewRepo.UpdateScore failed: %v", err)
		return nil, err
	}

	conversation.CurrentQuestionNumber++
	conversation.CurrentSubtopic = chatGPTResponse.NextSubtopic
	questionNumber++
	_, err = repo.UpdateConversationCurrents(conversationID, topicID, questionNumber, chatGPTResponse.NextSubtopic)
	if err != nil {
		log.Printf("UpdateConversationTopic error: %v", err)
		return nil, err
	}

	messagesQ2 := []Message{
		NewMessage(conversationID, topicID, questionNumber, Interviewer, chatGPTResponseString),
	}
	conversation.Topics[topicID].Questions[questionNumber] = NewQuestion(conversationID, topicID, questionNumber, chatGPTResponse.NextQuestion, messagesQ2)

	_, err = repo.AddQuestion(conversation.Topics[topicID].Questions[questionNumber])
	if err != nil {
		log.Printf("AddQuestion in CreateConversation err: %v", err)
		return nil, err
	}
	_, err = repo.AddMessage(conversationID, topicID, questionNumber, messagesQ2[0])
	if err != nil {
		log.Printf("AddMessage in CreateConversation err: %v", err)
		return nil, err
	}

	return conversation, nil
}

func AppendConversation(
	repo ConversationRepo,
	interviewRepo interview.InterviewRepo,
	openAI chatgpt.AIClient,
	interviewID,
	userID int,
	conversation *Conversation,
	message, prompt string) (*Conversation, error) {

	conversationID := conversation.ID
	topicID := conversation.CurrentTopic
	questionNumber := conversation.CurrentQuestionNumber

	if conversation.ID != conversationID {
		return nil, errors.New("conversation_id doesn't match with current interview")
	}

	messageUser := NewMessage(conversationID, topicID, questionNumber, User, message)
	_, err := repo.AddMessage(conversationID, topicID, questionNumber, messageUser)
	if err != nil {
		return nil, err
	}
	conversation.Topics[topicID].Questions[questionNumber].Messages = append(conversation.Topics[topicID].Questions[questionNumber].Messages, messageUser)

	chatGPTResponse, chatGPTResponseString, err := GetChatGPTResponses(conversation, openAI, interviewRepo, conversation.Topics[topicID].Questions[questionNumber].Prompt, message)
	if err != nil {
		log.Printf("getChatGPTResponses failed: %v", err)
		return nil, err
	}

	err = interviewRepo.UpdateScore(interviewID, chatGPTResponse.Score)
	if err != nil {
		log.Printf("interviewRepo.UpdateScore failed: %v", err)
		return nil, err
	}

	moveToNewTopic, incrementQuestion, isFinished, err := CheckConversationState(chatGPTResponse, conversation)
	if err != nil {
		log.Printf("CheckConversationState err: %v", err)
		return nil, err
	}

	if isFinished {
		conversation.CurrentTopic = 0
		conversation.CurrentSubtopic = "finished"
		conversation.CurrentQuestionNumber = 0

		err := interviewRepo.UpdateStatus(interviewID, userID, "finished")
		if err != nil {
			log.Printf("interviewRepo.UpdateStatus failed: %v", err)
			return nil, err
		}

		_, err = repo.UpdateConversationCurrents(conversationID, conversation.CurrentTopic, 0, conversation.CurrentSubtopic)
		if err != nil {
			log.Printf("UpdateConversationTopic error: %v", err)
			return nil, err
		}

		messageFinal := NewMessage(conversationID, conversation.CurrentTopic, questionNumber, Interviewer, chatGPTResponseString)
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
		conversation.CurrentQuestionNumber = resetQuestionNumber

		_, err := repo.UpdateConversationCurrents(conversationID, nextTopicID, resetQuestionNumber, chatGPTResponse.NextSubtopic)
		if err != nil {
			log.Printf("UpdateConversationTopic error: %v", err)
			return nil, err
		}

		topic := conversation.Topics[nextTopicID]
		topic.ConversationID = conversationID
		messages := []Message{
			NewMessage(conversationID, nextTopicID, resetQuestionNumber, Interviewer, chatGPTResponseString),
		}
		question := NewQuestion(conversationID, nextTopicID, resetQuestionNumber, chatGPTResponse.NextQuestion, messages)
		topic.Questions = make(map[int]*Question)
		topic.Questions[resetQuestionNumber] = question
		conversation.Topics[nextTopicID] = topic

		_, err = repo.AddQuestion(question)
		if err != nil {
			log.Printf("AddQuestion in AppendConversation err: %v", err)
		}
		_, err = repo.AddMessage(conversationID, nextTopicID, resetQuestionNumber, messages[0])
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
		conversation.Topics[topicID].Questions[questionNumber] = NewQuestion(conversationID, topicID, questionNumber, chatGPTResponse.NextQuestion, messages)
	}

	messageInterviewer := NewMessage(conversationID, topicID, questionNumber, Interviewer, chatGPTResponseString)
	conversation.Topics[topicID].Questions[questionNumber].Messages = append(conversation.Topics[topicID].Questions[questionNumber].Messages, messageInterviewer)

	_, err = repo.AddQuestion(conversation.Topics[topicID].Questions[questionNumber])
	if err != nil {
		log.Printf("AddQuestion in AppendConversation failed: %v", err)
		return nil, err
	}
	_, err = repo.AddMessage(conversationID, topicID, questionNumber, messageInterviewer)
	if err != nil {
		log.Printf("AddMessage in AppendConversation failed: %v", err)
		return nil, err
	}

	return conversation, nil
}

func GetConversation(repo ConversationRepo, interviewID int) (*Conversation, error) {
	conversation, err := repo.GetConversation(interviewID)
	if err != nil {
		return nil, err
	}

	conversation.Topics = ClonePredefinedTopics()

	questionsReturned, err := repo.GetQuestions(conversation)
	if err != nil {
		return nil, err
	}

	for _, question := range questionsReturned {
		topicID := question.TopicID
		topic := conversation.Topics[topicID]
		topic.ConversationID = conversation.ID

		if topic.Questions == nil {
			topic.Questions = make(map[int]*Question)
		}

		topic.Questions[question.QuestionNumber] = question

		messagesReturned, err := repo.GetMessages(conversation.ID, topicID, question.QuestionNumber)
		if err != nil {
			log.Printf("repo.GetMessages failed: %v\n", err)
			return nil, err
		}

		question.Messages = append(question.Messages, messagesReturned...)
		topic.Questions[question.QuestionNumber] = question
		conversation.Topics[topicID] = topic
	}

	return conversation, nil
}
