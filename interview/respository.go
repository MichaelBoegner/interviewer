package interview

import (
	"database/sql"
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
	query := `
    INSERT INTO interviews (user_id, length, number_questions, difficulty, status, score, language, questions)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    RETURNING id
    `

	var id int
	err := repo.DB.QueryRow(query,
		interview.UserId, interview.Length, interview.NumberQuestions,
		interview.Difficulty, interview.Status, interview.Score,
		interview.Language, interview.Questions).Scan(&id)

	if err != nil {
		return 0, err
	}

	// Return the generated id
	return id, nil
}
