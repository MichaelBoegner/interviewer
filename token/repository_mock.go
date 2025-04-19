package token

import "errors"

type MockRepo struct {
	failRepo bool
}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (m *MockRepo) AddRefreshToken(token *RefreshToken) error {
	if m.failRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) GetStoredRefreshToken(userID int) (string, error) {
	if m.failRepo {
		return "", errors.New("Mocked DB failure")
	}

	return "abc123", nil
}
