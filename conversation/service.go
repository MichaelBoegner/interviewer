package conversation

import "time"

func CreateConversation(repo ConversationRepo, interviewID int, messages map[string]string) (*Conversation, error) {
	now := time.Now()

	conversation := &Conversation{
		InterviewID: interviewID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	conversationID, err := repo.CreateConversation(conversation)
	if err != nil {
		return nil, err
	}

	conversation.ID = conversationID

	return conversation, nil
}
