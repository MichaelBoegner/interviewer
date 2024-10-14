package user

import "time"

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
	CreateUser(user *User) error
	GetPasswordandID(username string) (int, string, error)
	GetUser(user *User) (*User, error)
}
