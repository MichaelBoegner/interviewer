package interview

import (
	"errors"
	"time"
)

type MockRepo struct {
	failRepo bool
}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (m *MockRepo) CreateInterview(interview *Interview) (int, error) {
	if m.failRepo {
		return 0, errors.New("Mocked DB failure")
	}

	return 1, nil
}

func (m *MockRepo) GetInterview(interviewID int) (*Interview, error) {
	if m.failRepo {
		return nil, errors.New("Mocked DB failure")
	}

	interview := &Interview{
		Id:              1,
		UserId:          1,
		Length:          30,
		NumberQuestions: 2,
		Difficulty:      "easy",
		Status:          "running",
		Score:           0,
		Language:        "python",
		FirstQuestion:   "question1",
		Subtopic:        "None",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	return interview, nil
}
