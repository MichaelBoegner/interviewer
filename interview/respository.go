package interview

import (
	"database/sql"
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
	// fmt.Printf("CreateInterview firing: %v\n", interview)

	query := `
    INSERT INTO interviews (user_id, length, number_questions, difficulty, status, score, language, prompt, first_question)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    RETURNING id
    `

	var id int
	err := repo.DB.QueryRow(query,
		interview.UserId,
		interview.Length,
		interview.NumberQuestions,
		interview.Difficulty,
		interview.Status,
		interview.Score,
		interview.Language,
		interview.Prompt,
		interview.FirstQuestion).Scan(&id)

	if err != nil {
		return 0, err
	}

	// fmt.Printf("query fired in repo and id is: %v\n", id)
	return id, nil
}

func (repo *Repository) GetInterview(interviewID int) (*Interview, error) {
	query := `
	SELECT user_id, length, number_questions, difficulty, status, score, language, prompt, first_question
	FROM interviews
	WHERE id = $1
	`

	interview := &Interview{}
	err := repo.DB.QueryRow(query,
		interviewID).Scan(
		&interview.UserId,
		&interview.Length,
		&interview.NumberQuestions,
		&interview.Difficulty,
		&interview.Status,
		&interview.Score,
		&interview.Language,
		&interview.Prompt,
		&interview.FirstQuestion)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no interview found with id %d", interviewID)
		}
		return nil, err
	}

	return interview, nil
}
