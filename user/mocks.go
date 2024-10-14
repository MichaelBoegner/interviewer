package user

type MockRepo struct {
	Users map[int]User
}

func NewMockRepo() *MockRepo {
	return &MockRepo{
		Users: map[int]User{
			1: {ID: 1, Username: "testuser", Password: []byte("$2a$10$...")},
		},
	}
}

func (m *MockRepo) CreateUser(user *User) error {
	m.Users[0] = *user
	return nil
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
	return 1, "testuser", nil
}
