package token

type MockRepo struct{}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (m *MockRepo) AddRefreshToken(token *RefreshToken) error {
	return nil
}

func (m *MockRepo) GetStoredRefreshToken(userID int) (string, error) {
	return "9942443a086328dfaa867e0708426f94284d25700fa9df930261e341f0d8c671", nil
}
