package token

import "time"

type MockRepo struct {
	UserID       int
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (m *MockRepo) AddRefreshToken(token *RefreshToken) error {
	return nil
}

func (m *MockRepo) GetStoredRefreshToken(userID int) (string, error) {
	return "", nil
}
