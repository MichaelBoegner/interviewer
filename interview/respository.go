package interview

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

func (repo *Repository) CreateInterview(interview *Interview) (int, error) {
	questionsJSON, err := json.Marshal(interview.Questions)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal questions: %v", err)
	}

	query := `
    INSERT INTO interviews (user_id, length, number_questions, difficulty, status, score, language, questions)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    RETURNING id
    `

	var id int
	err = repo.DB.QueryRow(query,
		interview.UserId, interview.Length, interview.NumberQuestions,
		interview.Difficulty, interview.Status, interview.Score,
		interview.Language, questionsJSON).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}
