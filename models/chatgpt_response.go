package models

import "time"

type ChatGPTResponse struct {
	Topic        string    `json:"topic"`
	Subtopic     string    `json:"subtopic"`
	Question     string    `json:"question"`
	Score        int       `json:"score"`
	Feedback     string    `json:"feedback"`
	NextQuestion string    `json:"next_question"`
	NextTopic    string    `json:"next_topic"`
	NextSubtopic string    `json:"next_subtopic"`
	CreatedAt    time.Time `json:"created_at"`
}
