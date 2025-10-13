package mocks

type MockMailerService struct{}

func NewMockMailerService() *MockMailerService {
	return &MockMailerService{}
}

func (m *MockMailerService) SendPasswordReset(email, resetURL string) error {
	return nil
}

func (m *MockMailerService) SendVerificationEmail(email, verifyURL string) error {
	return nil
}

func (m *MockMailerService) SendWelcome(email string) error {
	return nil
}

func (m *MockMailerService) SendDeletionConfirmation(email string) error {
	return nil
}
