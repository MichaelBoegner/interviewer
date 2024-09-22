package user

type Users struct {
	Users map[int]User
}

type User struct {
	Id       int
	Username string
	Email    string
}
