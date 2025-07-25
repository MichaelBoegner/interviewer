package mocks

type MockMailer struct{}

func NewMockMailer() *MockMailer {
	mockMailer := &MockMailer{}

	return mockMailer
}

func (m *MockMailer) SendPasswordReset(email, resetURL string) error {
	return nil
}

func (m *MockMailer) SendVerificationEmail(email, verifyURL string) error {
	return nil
}

func (m *MockMailer) SendWelcome(email string) error {
	return nil
}

func (m *MockMailer) SendDeletionConfirmation(email string) error {
	return nil
}
