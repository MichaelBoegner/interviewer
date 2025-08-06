package interview

import (
	"database/sql"
	"fmt"
	"log"
	"time"
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
    INSERT INTO interviews (
	user_id, 
	length, 
	number_questions, 
	difficulty, 
	status, 
	score, 
	language, 
	prompt, 
	jd_summary,
	first_question, 
	subtopic,
	created_at,
	updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
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
		interview.JDSummary,
		interview.FirstQuestion,
		interview.Subtopic,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *Repository) LinkConversation(interviewID, conversationID int) error {
	query := `
		UPDATE interviews
		SET conversation_id = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := repo.DB.Exec(query, conversationID, time.Now().UTC(), interviewID)
	if err != nil {
		log.Printf("LinkConversation failed: %v", err)
		return err
	}

	return nil
}

func (repo *Repository) GetInterview(interviewID int) (*Interview, error) {

	query := `
	SELECT 
		id, 
		conversation_id,
		user_id, 
		length, 
		number_questions, 
		difficulty, 
		status, 
		score, 
		language, 
		prompt, 
		jd_summary,
		first_question, 
		subtopic,
		updated_at,
		created_at
	FROM interviews
	WHERE id = $1
	`

	interview := &Interview{}
	err := repo.DB.QueryRow(query,
		interviewID).Scan(
		&interview.Id,
		&interview.ConversationID,
		&interview.UserId,
		&interview.Length,
		&interview.NumberQuestions,
		&interview.Difficulty,
		&interview.Status,
		&interview.Score,
		&interview.Language,
		&interview.Prompt,
		&interview.JDSummary,
		&interview.FirstQuestion,
		&interview.Subtopic,
		&interview.UpdatedAt,
		&interview.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no interview found with id %d", interviewID)
		}
		return nil, err
	}

	return interview, nil
}

func (repo *Repository) GetInterviewSummariesByUserID(userID int) ([]Summary, error) {
	rows, err := repo.DB.Query(`
		SELECT id, created_at, score, status
		FROM interviews
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []Summary
	for rows.Next() {
		var summary Summary
		err := rows.Scan(&summary.ID, &summary.StartedAt, &summary.Score, &summary.Status)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func (repo *Repository) UpdateScore(interviewID, pointsEarned int) error {
	query := `
		UPDATE interviews
		SET
			number_questions_answered = number_questions_answered + 1,
			score_numerator = score_numerator + $1,
			score = ROUND((score_numerator + $1)::decimal / ((number_questions_answered + 1) * 10) * 100),
			updated_at = $2
		WHERE id = $3
	`
	_, err := repo.DB.Exec(query, pointsEarned, time.Now().UTC(), interviewID)
	return err
}

func (repo *Repository) UpdateStatus(interviewID, userID int, status string) error {
	query := `UPDATE interviews SET status = $1, updated_at = $2 WHERE id = $3 AND user_id = $4`
	_, err := repo.DB.Exec(query, status, time.Now().UTC(), interviewID, userID)
	return err
}
