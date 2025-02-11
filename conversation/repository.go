package conversation

import (
	"database/sql"
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
	var id int
	query := `SELECT interview_id
	FROM conversations
	WHERE interview_id = $1
	`
	err := repo.DB.QueryRow(query, interviewID).Scan(&id)

	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		log.Printf("Error querying conversation: %v\n", err)
		return false
	}

	return true
}

func (repo *Repository) GetConversation(interviewID int) (*Conversation, error) {
	conversation := &Conversation{}

	query := `SELECT id, interview_id, current_topic, current_question_number, created_at, updated_at
	FROM conversations
	WHERE interview_id = $1
	`
	err := repo.DB.QueryRow(query, interviewID).Scan(
		&conversation.ID,
		&conversation.InterviewID,
		&conversation.CurrentTopic,
		&conversation.CurrentQuestionNumber,
		&conversation.CreatedAt,
		&conversation.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		log.Printf("Error querying conversation: %v\n", err)
		return nil, err
	}

	return conversation, nil
}

func (repo *Repository) CreateConversation(conversation *Conversation) (int, error) {
	var id int
	query := `
		INSERT INTO conversations (interview_id, current_topic, current_question_number, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
		`

	err := repo.DB.QueryRow(query,
		conversation.InterviewID,
		conversation.CurrentTopic,
		conversation.CurrentQuestionNumber,
		conversation.CreatedAt,
		conversation.UpdatedAt,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, err
	} else if err != nil {
		log.Printf("Error querying conversation: %v\n", err)
		return 0, err
	}

	return id, nil
}

func (repo *Repository) CreateQuestion(conversation *Conversation, prompt string) (int, error) {
	var id int

	query := `
			INSERT INTO questions (conversation_id, topic_id, question_number, prompt, created_at) 
			VALUES ($1, $2, $3, $4, $5)
			RETURNING question_number
			`

	err := repo.DB.QueryRow(query,
		conversation.ID,
		conversation.CurrentTopic,
		1,
		prompt,
		time.Now(),
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, err
	} else if err != nil {
		log.Printf("Error querying conversation: %v\n", err)
		return 0, err
	}

	return id, nil
}

func (repo *Repository) AddQuestion(conversation *Conversation, questionNumber int, prompt string) (int, error) {
	var id int

	query := `
			INSERT INTO questions (conversation_id, topic_id, question_number, prompt, created_at) 
			VALUES ($1, $2, $3, $4, $5)
			RETURNING question_number
			`

	err := repo.DB.QueryRow(query,
		conversation.ID,
		conversation.CurrentTopic,
		questionNumber,
		prompt,
		time.Now(),
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, err
	} else if err != nil {
		log.Printf("Error querying conversation: %v\n", err)
		return 0, err
	}

	return id, nil
}

func (repo *Repository) GetQuestions(conversation *Conversation) ([]*Question, error) {
	var questions []*Question

	query := `
			SELECT conversation_id, topic_id, question_number, prompt, created_at
			FROM questions 
			WHERE conversation_id = ($1)
			`

	rows, err := repo.DB.Query(query, conversation.ID)
	if err != nil {
		log.Printf("Error querying conversation: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		question := &Question{}
		err := rows.Scan(
			&question.ConversationID,
			&question.TopicID,
			&question.QuestionNumber,
			&question.Prompt,
			&question.CreatedAt)
		if err != nil {
			log.Printf("Error scanning row: %v\n", err)
			return nil, err
		}
		questions = append(questions, question)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after iterating rows: %v\n", err)
		return nil, err
	}

	return questions, nil
}

func (repo *Repository) CreateMessages(conversation *Conversation, messages []Message) error {
	var id int
	for _, message := range messages {
		query := `
			INSERT INTO messages (conversation_id, question_number, author, content, created_at) 
			VALUES ($1, $2, $3, $4, $5) 
			RETURNING id
			`

		err := repo.DB.QueryRow(query,
			conversation.ID,
			message.QuestionNumber,
			message.Author,
			message.Content,
			time.Now(),
		).Scan(&id)

		if err == sql.ErrNoRows {
			return err
		} else if err != nil {
			log.Printf("Error querying conversation: %v\n", err)
			return err
		}
	}

	return nil
}

func (repo *Repository) AddMessage(conversationID, questionNumber int, message *Message) (int, error) {
	query := `
			INSERT INTO messages (conversation_id, question_number, author, content, created_at) 
			VALUES ($1, $2, $3, $4, $5) 
			RETURNING question_number
			`

	err := repo.DB.QueryRow(query,
		conversationID,
		questionNumber,
		message.Author,
		message.Content,
		time.Now(),
	).Scan(&questionNumber)
	if err == sql.ErrNoRows {
		return 0, err
	} else if err != nil {
		log.Printf("Error querying conversation: %v\n", err)
		return 0, err
	}

	return questionNumber, nil
}

func (repo *Repository) GetMessages(conversationID, questionNumber int) ([]Message, error) {
	messages := make([]Message, 0)

	query := `
			SELECT conversation_id, question_number, author, content, created_at
			FROM messages
			WHERE conversation_id = $1 and question_number = $2
			`

	rows, err := repo.DB.Query(query, conversationID, questionNumber)
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		log.Printf("Error querying conversation: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		err := rows.Scan(
			&message.ConversationID,
			&message.QuestionNumber,
			&message.Author,
			&message.Content,
			&message.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning message: %v\n", err)
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (repo *Repository) UpdateConversationCurrents(topicID, questionNumber, conversationID int) (int, error) {
	var id int

	query := `
			UPDATE conversations
			SET current_topic = $1, current_question_number = $2, updated_at = $3
			WHERE id = $4
			RETURNING id;
			`

	err := repo.DB.QueryRow(query,
		topicID,
		questionNumber,
		time.Now(),
		conversationID,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, err
	} else if err != nil {
		log.Printf("UpdateConversationTopic error: %v\n", err)
		return 0, err
	}

	return id, nil
}
