package conversation

import (
	"database/sql"
	"encoding/json"
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

func (repo *Repository) CreateConversation(conversation *Conversation) (int, error) {
	var id int
	query := `
		INSERT INTO conversations (interview_id, created_at, updated_at) 
		VALUES ($1, $2, $3)
		RETURNING id
		`

	err := repo.DB.QueryRow(query,
		conversation.InterviewID,
		conversation.CreatedAt,
		conversation.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *Repository) CreateTopics(conversation *Conversation) error {
	for _, topic := range conversation.Topics {
		topicName, err := json.Marshal(topic.Name)
		if err != nil {
			return err
		}

		query := `
			INSERT INTO topics (conversation_id, name) 
			VALUES ($1, $2)
			RETURNING id
			`

		repo.DB.QueryRow(query,
			conversation.ID,
			topicName,
		)
	}

	return nil
}

func (repo *Repository) CreateQuestion(conversation *Conversation) error {
	query := `
			INSERT INTO questions (topic_id, question_number, created_at) 
			VALUES ($1, $2, $3)
			`

	repo.DB.QueryRow(query,
		1,
		1,
		time.Now(),
	)

	return nil
}

func (repo *Repository) CreateMessages(author, content string) error {
	query := `
			INSERT INTO questions (question_id, author, content, create_at) 
			VALUES ($1, $2, $3, $4)
			`

	repo.DB.QueryRow(query,
		1,
		author,
		content,
		time.Now(),
	)

	return nil
}
