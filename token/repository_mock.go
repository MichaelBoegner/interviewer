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
		return errors.New("mocked DB failure")
	}

	return nil
}

func (m *MockRepo) GetStoredRefreshToken(userID int) (string, error) {
	if m.failRepo {
		return "", errors.New("mocked DB failure")
	}

	if userID != 1 {
		return "", errors.New("userID does not exist")
	}

	return "abc123", nil
}

func (m *MockRepo) DeleteRefreshToken(userID int) error {
	if m.failRepo {
		return errors.New("mocked DB failure")
	}

	return nil
}
