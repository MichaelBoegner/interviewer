package user

import (
	"errors"
	"time"
)

type Users struct {
	Users map[int]User
}

type User struct {
	ID        int
	Username  string
	Email     string
	Password  []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepo interface {
	CreateUser(user *User) (int, error)
	GetPasswordandID(username string) (int, string, error)
	GetUser(user *User) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdatePasswordByEmail(email string, password []byte) error
}

var (
	ErrDuplicateEmail    = errors.New("duplicate email")
	ErrDuplicateUsername = errors.New("duplicate username")
	ErrDuplicateUser     = errors.New("duplicate user")
)
