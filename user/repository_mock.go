package user

import (
	"errors"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type MockRepo struct {
	Users              map[int]User
	FailRepo           bool
	FailGetUserByEmail bool
	FailAddCredits     bool
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
		Users:              map[int]User{},
		FailGetUserByEmail: false,
	}
}

func (m *MockRepo) CreateUser(user *User) (int, error) {
	if m.FailRepo {
		return 0, errors.New("Mocked DB failure")
	}

	return 1, nil
}

func (m *MockRepo) MarkUserDeleted(userID int) error {
	if m.FailRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) GetUser(userID int) (*User, error) {
	if m.FailRepo {
		return nil, errors.New("Mocked DB failure")
	}

	mockUser := &User{
		ID:            1,
		Username:      "test",
		Password:      PasswordHashed,
		Email:         "test@test.com",
		AccountStatus: "active",
	}

	return mockUser, nil
}

func (m *MockRepo) GetPasswordandID(username string) (int, string, error) {
	if m.FailRepo {
		return 0, "", errors.New("Mocked DB failure")
	}

	return 1, string(PasswordHashed), nil
}

func (m *MockRepo) GetUserByEmail(email string) (*User, error) {
	if m.FailGetUserByEmail {
		return nil, errors.New("Mocked GetUserByEmail failure")
	}
	if m.FailRepo {
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
	if m.FailRepo {
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
	if m.FailRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) AddCredits(userID, credits int, creditType string) error {
	if m.FailAddCredits {
		return errors.New("Mocked AddCredits failure")
	}
	if m.FailRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) UpdateSubscriptionData(userID int, status, tier, subscriptionID string, startsAt, endsAt time.Time) error {
	if m.FailRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) UpdateSubscriptionStatusData(userID int, status string) error {
	if m.FailRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) HasActiveOrCancelledSubscription(email string) (bool, error) {
	if m.FailRepo {
		return false, errors.New("Mocked DB failure")
	}

	return true, nil
}
