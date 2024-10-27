package conversation

type MockRepo struct {
}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (repo *MockRepo) CheckForConversation(interviewID int) bool {
	return true
}

func (repo *MockRepo) CreateConversation(conversation *Conversation) (int, error) {
	return 1, nil
}

func (repo *MockRepo) CreateTopics(Conversation *Conversation) error {
	return nil
}

func (repo *MockRepo) CreateQuestion(conversation *Conversation) error {
	return nil
}

func (repo *MockRepo) CreateMessages(author, content string) error {
	return nil
}
