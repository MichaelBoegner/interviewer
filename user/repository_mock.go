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

var (
	PasswordHashed []byte
	err            error
)

func NewMockRepo() *MockRepo {
	PasswordHashed, err = bcrypt.GenerateFromPassword([]byte("test"), bcrypt.MinCost)
	if err != nil {
		log.Printf("GenerateFromPassword in NewMockRepo() failed: %v", err)
	}

	return &MockRepo{
		Users: map[int]User{},
	}
}

func (m *MockRepo) CreateUser(user *User) (int, error) {
	if m.failRepo {
		return 0, errors.New("Mocked DB failure")
	}

	return 1, nil
}

func (m *MockRepo) GetUser(user *User) (*User, error) {
	if m.failRepo {
		return nil, errors.New("Mocked DB failure")
	}

	mockUser := &User{
		ID:       1,
		Username: "test",
		Password: PasswordHashed,
		Email:    "test@test.com",
	}

	return mockUser, nil
}

func (m *MockRepo) GetPasswordandID(username string) (int, string, error) {
	if m.failRepo {
		return 0, "", errors.New("Mocked DB failure")
	}

	return 1, string(PasswordHashed), nil
}
