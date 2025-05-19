package billing

type MockRepo struct {
	failRepo bool
}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (m *MockRepo) LogCreditTransaction(tx CreditTransaction) error {
	return nil
}

func (m *MockRepo) HasWebhookBeenProcessed(id string) (bool, error) {
	return false, nil
}

func (m *MockRepo) MarkWebhookProcessed(id string, event string) error {
	return nil
}
