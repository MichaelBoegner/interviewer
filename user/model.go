package user

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Users struct {
	Users map[int]User
}

type User struct {
	ID                    int
	Username              string
	Email                 string
	Password              []byte
	SubscriptionTier      string
	SubscriptionStatus    string
	SubscriptionStartDate *time.Time
	SubscriptionEndDate   *time.Time
	SubscriptionID        string
	IndividualCredits     int
	SubscriptionCredits   int
	AccountStatus         string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type EmailClaims struct {
	Email        string `json:"email"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	Purpose      string `json:"purpose"`
	jwt.RegisteredClaims
}

type UserRepo interface {
	CreateUser(user *User) (int, error)
	MarkUserDeleted(userID int) error
	GetPasswordandID(email string) (int, string, error)
	GetUser(userID int) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByCustomerID(customerID string) (*User, error)
	UpdatePasswordByEmail(email string, password []byte) error
	AddCredits(userID, credits int, creditType string) error
	UpdateSubscriptionData(userID int, status, tier, subscriptionID string, startsAt, endsAt time.Time) error
	UpdateSubscriptionStatusData(userID int, status string) error
	HasActiveOrCancelledSubscription(email string) (bool, error)
}

var (
	ErrDuplicateEmail    = errors.New("duplicate email")
	ErrDuplicateUsername = errors.New("duplicate username")
	ErrDuplicateUser     = errors.New("duplicate user")
	ErrAccountDeleted    = errors.New("Account is no longer active")
)
