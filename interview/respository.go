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
	query := `
    INSERT INTO interviews (user_id, length, number_questions, difficulty, status, score, language, prompt, first_question, subtopic)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
		interview.FirstQuestion,
		interview.Subtopic,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *Repository) GetInterview(interviewID int) (*Interview, error) {

	query := `
	SELECT id, user_id, length, number_questions, difficulty, status, score, language, prompt, first_question, subtopic
	FROM interviews
	WHERE id = $1
	`

	interview := &Interview{}
	err := repo.DB.QueryRow(query,
		interviewID).Scan(
		&interview.Id,
		&interview.UserId,
		&interview.Length,
		&interview.NumberQuestions,
		&interview.Difficulty,
		&interview.Status,
		&interview.Score,
		&interview.Language,
		&interview.Prompt,
		&interview.FirstQuestion,
		&interview.Subtopic)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no interview found with id %d", interviewID)
		}
		return nil, err
	}

	return interview, nil
}

func (r *Repository) GetInterviewSummariesByUserID(userID int) ([]Summary, error) {
	rows, err := r.DB.Query(`
		SELECT id, created_at, score
		FROM interviews
		WHERE user_id = $1
		ORDER BY started_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []Summary
	for rows.Next() {
		var summary Summary
		err := rows.Scan(&summary.ID, &summary.StartedAt, &summary.Score)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}
