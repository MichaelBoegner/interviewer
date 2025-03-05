package models

import "time"

type ChatGPTResponse struct {
	Topic        string    `json:"topic"`
	Subtopic     string    `json:"subtopic"`
	Question     string    `json:"question"`
	Score        int       `json:"score"`
	Feedback     string    `json:"feedback"`
	NextQuestion string    `json:"next_question"`
	CreatedAt    time.Time `json:"created_at"`
}
