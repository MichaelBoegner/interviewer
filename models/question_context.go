package models

import "time"

type QuestionContext struct {
	Topic             string    `json:"topic"`
	Subtopic          string    `json:"subtopic"`
	Question          string    `json:"question"`
	Score             int       `json:"score"`
	Feedback          string    `json:"feedback"`
	NextQuestion      string    `json:"next_question"`
	MoveToNewSubtopic bool      `json:"move_to_new_subtopic"`
	MoveToNewTopic    bool      `json:"move_to_new_topic"`
	CreatedAt         time.Time `json:"created_at"`
}
