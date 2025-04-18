package user

import (
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type MockRepo struct {
	Users    map[int]User
	failRepo bool
}

func NewMockRepo() *MockRepo {
	return &MockRepo{
		Users: map[int]User{
			1: {ID: 1, Username: "testuser", Password: []byte("$2a$10$...")},
		},
	}
}

func (m *MockRepo) CreateUser(user *User) (int, error) {
	if m.failRepo {
		return 0, errors.New("Mocked DB failure")
	}

	m.Users[0] = *user
	return 1, nil
}

func (m *MockRepo) GetUser(user *User) (*User, error) {
	mockUser := &User{
		ID:       user.ID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	return mockUser, nil
}

func (m *MockRepo) GetPasswordandID(username string) (int, string, error) {
	passwordHashed, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return 0, "", err
	}

	passwordString := string(passwordHashed)
	return 1, passwordString, nil
}
