package interview

import (
	"time"
)

type Interview struct {
	Id              int
	UserId          int
	Length          int
	NumberQuestions int
	Difficulty      string
	Status          string
	Score           int
	Language        string
	Questions       map[int]string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
