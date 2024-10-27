package conversation

import (
	"time"
)

func CheckForConversation(repo ConversationRepo, interviewID int) bool {
	return repo.CheckForConversation(interviewID)
}

func CreateConversation(repo ConversationRepo, interviewID int, message *Message) (*Conversation, error) {
	now := time.Now()
	conversation := &Conversation{
		InterviewID: interviewID,
		Topics:      PredefinedTopics,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	topic := conversation.Topics[1]
	topic.Questions = make(map[int]Question)
	question := topic.Questions[1]
	question.ID = 1

	message.ID = 1
	message.CreatedAt = time.Now()

	question.Messages = make([]Message, 0)
	question.Messages = append(question.Messages, *message)

	conversation.Topics[1] = topic
	conversation.Topics[1].Questions[1] = question

	conversationID, err := repo.CreateConversation(conversation)
	if err != nil {
		return nil, err
	}
	conversation.ID = conversationID
	err = repo.CreateTopics(conversation)
	if err != nil {
		return nil, err
	}

	err = repo.CreateQuestion(conversation)
	if err != nil {
		return nil, err
	}

	err = repo.CreateMessages(string(message.Author), message.Content)
	if err != nil {
		return nil, err
	}

	return conversation, nil
}
