package conversation

import (
	"errors"
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
