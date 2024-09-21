package interview

import (
	"fmt"
)

func StartInterview(userId, length, numberQuestions int, difficulty string) (*Interview, error) {
	questions := map[int]string{1: "What is the speed of a swallow", 2: "What is your favorite color?"}

	interview := &Interview{
		UserId:          userId,
		Length:          length,
		NumberQuestions: numberQuestions,
		Difficulty:      difficulty,
		Questions:       questions,
	}

	fmt.Printf("Interview: %v", interview)
	return interview, nil
}
