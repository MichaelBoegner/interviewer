package conversation

import (
	"database/sql"
	"fmt"
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
	fmt.Printf("Checkforconversations firing: %v\n", interviewID)
	var id int
	query := `SELECT interview_id
	FROM conversations
	WHERE interview_id = $1
	`
	err := repo.DB.QueryRow(query, interviewID).Scan(&id)

	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		fmt.Printf("Error querying conversation: %v\n", err)
		return false
	}

	return true
}

func (repo *Repository) GetConversation(interviewID int) (*Conversation, error) {
	conversation := &Conversation{}

	query := `SELECT id, interview_id, created_at, updated_at
	FROM conversations
	WHERE interview_id = $1
	`
	err := repo.DB.QueryRow(query, interviewID).Scan(&conversation.ID, &conversation.InterviewID, &conversation.CreatedAt, &conversation.UpdatedAt)
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
	for _, message := range messages {
		query := `
			INSERT INTO messages (conversation_id, question_id, author, content, created_at) 
			VALUES ($1, $2, $3, $4, $5) 
			RETURNING id
			`

		repo.DB.QueryRow(query,
			conversation.ID,
			message.QuestionID,
			message.Author,
			message.Content,
			time.Now(),
		)
	}

	return nil
}

func (repo *Repository) AddMessage(questionID int, message *Message) (int, error) {
	var id int
	query := `
			INSERT INTO questions (question_id, author, content, created_at) 
			VALUES ($1, $2, $3, $4, $5) 
			RETURNING id
			`

	err := repo.DB.QueryRow(query,
		questionID,
		message.Author,
		message.Content,
		time.Now(),
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
