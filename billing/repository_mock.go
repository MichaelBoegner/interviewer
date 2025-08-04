package billing

import "errors"

type MockRepo struct {
	FailLogCreditTransaction bool
}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (m *MockRepo) LogCreditTransaction(tx CreditTransaction) error {
	if m.FailLogCreditTransaction {
		return errors.New("Mocked LogCreditTransaction failure")
	}
	return nil
}

func (m *MockRepo) HasWebhookBeenProcessed(id string) (bool, error) {
	return false, nil
}

func (m *MockRepo) MarkWebhookProcessed(id string, event string) error {
	return nil
}
