package user

import (
	"errors"
	"log"
	"time"

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

func (m *MockRepo) GetUser(userID int) (*User, error) {
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

func (m *MockRepo) GetUserByEmail(email string) (*User, error) {
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

func (m *MockRepo) GetUserByCustomerID(customerID string) (*User, error) {
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

func (m *MockRepo) UpdatePasswordByEmail(email string, password []byte) error {
	if m.failRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) AddCredits(userID, credits int, creditType string) error {
	if m.failRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) UpdateSubscriptionData(userID int, status, tier string, startsAt, endsAt time.Time) error {
	if m.failRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) UpdateSubscriptionStatusData(userID int, status string) error {
	if m.failRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}
