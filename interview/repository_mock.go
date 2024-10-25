package interview

import (
	"encoding/json"
	"fmt"
	"time"
)

type MockRepo struct {
	Id              int
	UserId          int
	Length          int
	NumberQuestions int
	Difficulty      string
	Status          string
	Score           int
	Language        string
	Questions       map[string]string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (repo *MockRepo) CreateInterview(interview *Interview) (int, error) {
	_, err := json.Marshal(interview.Questions)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal questions: %v", err)
	}

	id := 1
	// Return the generated id
	return id, nil
}
