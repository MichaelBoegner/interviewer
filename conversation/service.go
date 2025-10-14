package conversation

import (
	"errors"
	"log"
)

func (c *ConversationService) CheckForConversation(interviewID int) (bool, error) {
	return c.ConversationRepo.CheckForConversation(interviewID)
}

func (c *ConversationService) CreateEmptyConversation(interviewID int, subTopic string) (int, error) {
	conversation := &Conversation{
		Topics:                ClonePredefinedTopics(),
		CurrentTopic:          1,
		CurrentSubtopic:       subTopic,
		CurrentQuestionNumber: 1,
	}

	conversationID, err := c.ConversationRepo.CreateConversation(interviewID, conversation)
	if err != nil {
		log.Printf("CreateConversation failed: %v", err)
		return 0, err
	}

	return conversationID, nil
}

func (c *ConversationService) CreateConversation(
	conversation *Conversation,
	interviewID int,
	prompt,
	firstQuestion,
	subtopic,
	message string) (*Conversation, error) {

	conversationID := conversation.ID
	topicID := conversation.CurrentTopic
	questionNumber := conversation.CurrentQuestionNumber

	_, err := c.ConversationRepo.CreateQuestion(conversation, firstQuestion)
	if err != nil {
		log.Printf("CreateQuestion failed: %v", err)
		return nil, err
	}

	topic := conversation.Topics[topicID]
	topic.ConversationID = conversationID
	messages := []Message{
		NewMessage(conversationID, topicID, questionNumber, System, prompt),
		NewMessage(conversationID, topicID, questionNumber, Interviewer, firstQuestion),
		NewMessage(conversationID, topicID, questionNumber, User, message),
	}
	topic.Questions = make(map[int]*Question)
	topic.Questions[questionNumber] = NewQuestion(conversationID, topicID, questionNumber, firstQuestion, messages)

	err = c.ConversationRepo.CreateMessages(conversation, messages)
	if err != nil {
		log.Printf("repo.CreateMessages failed: %v", err)
		return nil, err
	}

	chatGPTResponse, chatGPTResponseString, err := GetChatGPTResponses(conversation, c.AIService, c.InterviewRepo)
	if err != nil {
		log.Printf("getChatGPTResponses failed: %v", err)
		return nil, err
	}

	err = c.InterviewRepo.UpdateScore(interviewID, chatGPTResponse.Score)
	if err != nil {
		log.Printf("interviewRepo.UpdateScore failed: %v", err)
		return nil, err
	}

	conversation.CurrentQuestionNumber++
	conversation.CurrentSubtopic = chatGPTResponse.NextSubtopic
	questionNumber++
	_, err = c.ConversationRepo.UpdateConversationCurrents(conversationID, topicID, questionNumber, chatGPTResponse.NextSubtopic)
	if err != nil {
		log.Printf("UpdateConversationTopic error: %v", err)
		return nil, err
	}

	messagesQ2 := []Message{
		NewMessage(conversationID, topicID, questionNumber, Interviewer, chatGPTResponseString),
	}
	conversation.Topics[topicID].Questions[questionNumber] = NewQuestion(conversationID, topicID, questionNumber, chatGPTResponse.NextQuestion, messagesQ2)

	_, err = c.ConversationRepo.AddQuestion(conversation.Topics[topicID].Questions[questionNumber])
	if err != nil {
		log.Printf("AddQuestion in CreateConversation err: %v", err)
		return nil, err
	}
	_, err = c.ConversationRepo.AddMessage(conversationID, topicID, questionNumber, messagesQ2[0])
	if err != nil {
		log.Printf("AddMessage in CreateConversation err: %v", err)
		return nil, err
	}

	return conversation, nil
}

func (c *ConversationService) AppendConversation(
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
	_, err := c.ConversationRepo.AddMessage(conversationID, topicID, questionNumber, messageUser)
	if err != nil {
		return nil, err
	}
	conversation.Topics[topicID].Questions[questionNumber].Messages = append(conversation.Topics[topicID].Questions[questionNumber].Messages, messageUser)

	chatGPTResponse, chatGPTResponseString, err := GetChatGPTResponses(conversation, c.AIService, c.InterviewRepo)
	if err != nil {
		log.Printf("getChatGPTResponses failed: %v", err)
		return nil, err
	}

	err = c.InterviewRepo.UpdateScore(interviewID, chatGPTResponse.Score)
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

		err := c.InterviewRepo.UpdateStatus(interviewID, userID, "finished")
		if err != nil {
			log.Printf("interviewRepo.UpdateStatus failed: %v", err)
			return nil, err
		}

		_, err = c.ConversationRepo.UpdateConversationCurrents(conversationID, conversation.CurrentTopic, 0, conversation.CurrentSubtopic)
		if err != nil {
			log.Printf("UpdateConversationTopic error: %v", err)
			return nil, err
		}

		messageFinal := NewMessage(conversationID, topicID, questionNumber, Interviewer, chatGPTResponseString)
		_, err = c.ConversationRepo.AddMessage(conversationID, topicID, questionNumber, messageFinal)
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

		_, err := c.ConversationRepo.UpdateConversationCurrents(conversationID, nextTopicID, resetQuestionNumber, chatGPTResponse.NextSubtopic)
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

		_, err = c.ConversationRepo.AddQuestion(question)
		if err != nil {
			log.Printf("AddQuestion in AppendConversation err: %v", err)
		}
		_, err = c.ConversationRepo.AddMessage(conversationID, nextTopicID, resetQuestionNumber, messages[0])
		if err != nil {
			return nil, err
		}

		return conversation, nil
	}

	if incrementQuestion {
		conversation.CurrentQuestionNumber++
		questionNumber++
		_, err := c.ConversationRepo.UpdateConversationCurrents(conversationID, topicID, questionNumber, chatGPTResponse.NextSubtopic)
		if err != nil {
			log.Printf("UpdateConversationTopic error: %v", err)
			return nil, err
		}
		messages := []Message{}
		conversation.Topics[topicID].Questions[questionNumber] = NewQuestion(conversationID, topicID, questionNumber, chatGPTResponse.NextQuestion, messages)
	}

	messageInterviewer := NewMessage(conversationID, topicID, questionNumber, Interviewer, chatGPTResponseString)
	conversation.Topics[topicID].Questions[questionNumber].Messages = append(conversation.Topics[topicID].Questions[questionNumber].Messages, messageInterviewer)

	_, err = c.ConversationRepo.AddQuestion(conversation.Topics[topicID].Questions[questionNumber])
	if err != nil {
		log.Printf("AddQuestion in AppendConversation failed: %v", err)
		return nil, err
	}
	_, err = c.ConversationRepo.AddMessage(conversationID, topicID, questionNumber, messageInterviewer)
	if err != nil {
		log.Printf("AddMessage in AppendConversation failed: %v", err)
		return nil, err
	}

	return conversation, nil
}

func (c *ConversationService) GetConversation(interviewID int) (*Conversation, error) {
	conversation, err := c.ConversationRepo.GetConversation(interviewID)
	if err != nil {
		return nil, err
	}

	conversation.Topics = ClonePredefinedTopics()

	questionsReturned, err := c.ConversationRepo.GetQuestions(conversation)
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

		messagesReturned, err := c.ConversationRepo.GetMessages(conversation.ID, topicID, question.QuestionNumber)
		if err != nil {
			log.Printf("c.ConversationRepo.GetMessages failed: %v\n", err)
			return nil, err
		}

		question.Messages = append(question.Messages, messagesReturned...)
		topic.Questions[question.QuestionNumber] = question
	}

	return conversation, nil
}
