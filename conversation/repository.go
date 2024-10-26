package conversation

import (
	"database/sql"
	"encoding/json"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

func (repo *Repository) CreateConversation(conversation *Conversation) (int, error) {
	topicsJSON, err := json.Marshal(conversation.Topics)
	if err != nil {
		return 0, err
	}

	var id int
	query := `
		INSERT INTO conversations (interview_id, topics, created_at, updated_at) 
		VALUES ($1, $2, $3, $4)
		RETURNING id
		`

	err = repo.DB.QueryRow(query,
		conversation.InterviewID,
		topicsJSON,
		conversation.CreatedAt,
		conversation.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
