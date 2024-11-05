package conversation

type MockRepo struct{}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (repo *MockRepo) CheckForConversation(interviewID int) bool {
	return true
}

func (repo *MockRepo) GetConversation(interviewID int) (*Conversation, error) {
	return nil, nil
}

func (repo *MockRepo) CreateConversation(conversation *Conversation) (int, error) {
	return 1, nil
}

func (repo *MockRepo) CreateQuestion(conversation *Conversation) (int, error) {
	return 1, nil
}

func (repo *MockRepo) CreateMessages(conversation *Conversation, messages []Message) error {
	return nil
}

func (repo *MockRepo) AddMessage(questionID int, message *Message) (int, error) {
	return 3, nil
}
