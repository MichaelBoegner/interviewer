package interview

import (
	"time"
)

type MockRepo struct{}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (repo *MockRepo) CreateInterview(interview *Interview) (int, error) {
	id := 1
	return id, nil
}

func (repo *MockRepo) GetInterview(interviewID int) (*Interview, error) {
	interview := &Interview{
		Id:              1,
		UserId:          1,
		Length:          30,
		NumberQuestions: 5,
		Difficulty:      "easy",
		Status:          "running",
		Score:           0,
		Language:        "python",
		FirstQuestion:   "What is the flight speed of an unladdened swallow?",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	return interview, nil
}
