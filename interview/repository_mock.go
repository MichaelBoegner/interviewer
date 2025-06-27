package interview

import (
	"errors"
	"time"
)

type MockRepo struct {
	FailRepo bool
}

func NewMockRepo() *MockRepo {
	return &MockRepo{}
}

func (m *MockRepo) CreateInterview(interview *Interview) (int, error) {
	if m.FailRepo {
		return 0, errors.New("Mocked DB failure")
	}

	return 1, nil
}

func (m *MockRepo) GetInterview(interviewID int) (*Interview, error) {
	if m.FailRepo {
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

func (m *MockRepo) GetInterviewSummariesByUserID(userID int) ([]Summary, error) {
	if m.FailRepo {
		return nil, errors.New("Mocked DB failure")
	}

	var summaries []Summary
	score := 100
	summary := Summary{
		ID:        1,
		StartedAt: time.Now().UTC(),
		Score:     &score,
	}

	summaries = append(summaries, summary)

	return summaries, nil
}

func (m *MockRepo) UpdateScore(interviewID, pointsEarned int) error {
	if m.FailRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) UpdateStatus(interviewID, userID int, status string) error {
	if m.FailRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) LinkConversation(interviewID, conversationID int) error {
	if m.FailRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}

func (m *MockRepo) UpdateCreatedInterview(interview *Interview) error {
	if m.FailRepo {
		return errors.New("Mocked DB failure")
	}

	return nil
}
