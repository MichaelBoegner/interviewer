package conversation

import (
	"database/sql"
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

func (repo *Repository) CheckForConversation(interviewID int) bool {
	var id int
	query := `SELECT interview_id,
	FROM conversations
	WHERE interview_id = $1
	RETURNING interview_id
	`
	err := repo.DB.QueryRow(query).Scan(&id)

	return err != nil
}

func (repo *Repository) GetConversation(interviewID int) (*Conversation, error) {
	conversation := &Conversation{}

	query := `SELECT interview_id,
	FROM conversations
	WHERE interview_id = $1
	RETURNING id, interview_id, created_at, updated_at
	`
	err := repo.DB.QueryRow(query).Scan(&conversation)
	if err != nil {
		return nil, err
	}

	return conversation, nil
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

func (repo *Repository) CreateQuestion(conversation *Conversation) (int, error) {
	var id int

	query := `
			INSERT INTO questions (conversation_id, topic_id, question_number, created_at) 
			VALUES ($1, $2, $3, $4)
			RETURNING id
			`

	repo.DB.QueryRow(query,
		conversation.ID,
		1,
		1,
		time.Now(),
	).Scan(&id)

	return id, nil
}

func (repo *Repository) CreateMessages(conversation *Conversation, messages []Message) error {
	var id int
	for _, message := range messages {
		query := `
			INSERT INTO questions (conversation_id, question_id, author, content, create_at) 
			VALUES ($1, $2, $3, $4, $5) 
			RETURNING id
			`

		repo.DB.QueryRow(query,
			conversation.ID,
			id,
			message.Author,
			message.Content,
			time.Now(),
		)

		id += 1
	}

	return nil
}
