package user

import "time"

type Users struct {
	Users map[int]User
}

type User struct {
	Id        int
	Username  string
	Email     string
	Password  []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}
