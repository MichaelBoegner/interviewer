package conversation

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
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		fmt.Printf("Error querying conversation: %v\n", err)
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
	if err == sql.ErrNoRows {
		return 0, err
	} else if err != nil {
		fmt.Printf("Error querying conversation: %v\n", err)
		return 0, err
	}

	return id, nil
}

func (repo *Repository) CreateQuestion(conversation *Conversation, prompt string) (int, error) {
	var id int

	query := `
			INSERT INTO questions (conversation_id, topic_id, question_number, prompt, created_at) 
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
			`

	err := repo.DB.QueryRow(query,
		conversation.ID,
		1,
		1,
		prompt,
		time.Now(),
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, err
	} else if err != nil {
		fmt.Printf("Error querying conversation: %v\n", err)
		return 0, err
	}

	return id, nil
}

func (repo *Repository) GetQuestion(conversation *Conversation) (*Question, error) {
	question := &Question{}

	query := `
			SELECT id, conversation_id, topic_id, question_number, prompt, created_at
			FROM questions 
			WHERE conversation_id = ($1)
			`

	err := repo.DB.QueryRow(query, conversation.ID).Scan(
		&question.ID,
		&question.ConversationID,
		&question.TopicID,
		&question.QuestionNumber,
		&question.Prompt,
		&question.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		log.Printf("Error querying conversation: %v\n", err)
		return nil, err
	}

	return question, nil
}

func (repo *Repository) CreateMessages(conversation *Conversation, messages []Message) error {
	var id int
	for _, message := range messages {
		query := `
			INSERT INTO messages (conversation_id, question_id, author, content, created_at) 
			VALUES ($1, $2, $3, $4, $5) 
			RETURNING id
			`

		err := repo.DB.QueryRow(query,
			conversation.ID,
			message.QuestionID,
			message.Author,
			message.Content,
			time.Now(),
		).Scan(&id)

		if err == sql.ErrNoRows {
			return err
		} else if err != nil {
			fmt.Printf("Error querying conversation: %v\n", err)
			return err
		}
	}

	return nil
}

func (repo *Repository) AddMessage(questionID int, message *Message) (int, error) {
	var id int
	query := `
			INSERT INTO messages (question_id, author, content, created_at) 
			VALUES ($1, $2, $3, $4) 
			RETURNING id
			`

	err := repo.DB.QueryRow(query,
		questionID,
		message.Author,
		message.Content,
		time.Now(),
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, err
	} else if err != nil {
		fmt.Printf("Error querying conversation: %v\n", err)
		return 0, err
	}

	return id, nil
}

func (repo *Repository) GetMessages(questionID int) ([]Message, error) {
	messages := make([]Message, 0)

	query := `
			SELECT id, question_id, author, content, created_at
			FROM messages
			WHERE question_id = $1
			`

	rows, err := repo.DB.Query(query, questionID)
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		fmt.Printf("Error querying conversation: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		err := rows.Scan(
			&message.ID,
			&message.QuestionID,
			&message.Author,
			&message.Content,
			&message.CreatedAt,
		)
		if err != nil {
			fmt.Printf("Error scanning message: %v\n", err)
			return nil, err
		}
		messages = append(messages, message)
	}

	fmt.Printf("messages: %v\n", messages)

	return messages, nil
}
