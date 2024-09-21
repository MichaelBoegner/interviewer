package interview

import (
	"fmt"
	"time"
)

func StartInterview(userId, length, numberQuestions int, difficulty string) (*Interview, error) {
	questions := map[int]string{1: "What is the speed of a swallow", 2: "What is your favorite color?"}
	now := time.Now()

	interview := &Interview{
		UserId:          userId,
		Length:          length,
		NumberQuestions: numberQuestions,
		Difficulty:      difficulty,
		Status:          "Running",
		Score:           100,
		Language:        "Python",
		Questions:       questions,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	fmt.Printf("Interview: %v", interview)
	return interview, nil
}
