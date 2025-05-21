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

func (repo *Repository) CheckForConversation(interviewID int) (bool, error) {
	var id int
	query := `SELECT interview_id
	FROM conversations
	WHERE interview_id = $1
	`
	err := repo.DB.QueryRow(query, interviewID).Scan(&id)

	if err == sql.ErrNoRows {
		return false, err
	} else if err != nil {
		log.Printf("Error querying conversation: %v\n", err)
		return false, err
	}

	return true, nil
}

func (repo *Repository) CreateConversation(conversation *Conversation) (int, error) {
	var id int
	query := `
		INSERT INTO conversations (interview_id, current_topic, current_subtopic, current_question_number, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
		`

	err := repo.DB.QueryRow(query,
		conversation.InterviewID,
		conversation.CurrentTopic,
		conversation.CurrentSubtopic,
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

func (repo *Repository) GetConversation(interviewID int) (*Conversation, error) {
	conversation := &Conversation{}

	query := `SELECT id, interview_id, current_topic, current_subtopic, current_question_number, created_at, updated_at
	FROM conversations
	WHERE interview_id = $1
	`
	err := repo.DB.QueryRow(query, interviewID).Scan(
		&conversation.ID,
		&conversation.InterviewID,
		&conversation.CurrentTopic,
		&conversation.CurrentSubtopic,
		&conversation.CurrentQuestionNumber,
		&conversation.CreatedAt,
		&conversation.UpdatedAt)
	if err == sql.ErrNoRows {
		log.Printf("repo.GetConversation returned 0 rows: %v\n", err)
		return nil, err
	} else if err != nil {
		log.Printf("repo.GetConversation failed: %v\n", err)
		return nil, err
	}

	return conversation, nil
}

func (repo *Repository) UpdateConversationCurrents(conversationID, topicID, currentQuestionNumber int, subtopic string) (int, error) {
	var id int

	query := `
			UPDATE conversations
			SET current_topic = $1, current_subtopic = $2, current_question_number = $3, updated_at = $4
			WHERE id = $5
			RETURNING id;
			`

	err := repo.DB.QueryRow(query,
		topicID,
		subtopic,
		currentQuestionNumber,
		time.Now().UTC(),
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

func (repo *Repository) CreateQuestion(conversation *Conversation, prompt string) (int, error) {
	var questionNumber int

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
		time.Now().UTC(),
	).Scan(&questionNumber)
	if err == sql.ErrNoRows {
		return 0, err
	} else if err != nil {
		log.Printf("Error querying conversation: %v\n", err)
		return 0, err
	}

	return questionNumber, nil
}

func (repo *Repository) AddQuestion(question *Question) (int, error) {
	var id int

	query := `
			INSERT INTO questions (conversation_id, topic_id, question_number, prompt, created_at) 
			VALUES ($1, $2, $3, $4, $5)
			RETURNING question_number
			`

	err := repo.DB.QueryRow(query,
		question.ConversationID,
		question.TopicID,
		question.QuestionNumber,
		question.Prompt,
		time.Now().UTC(),
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
			INSERT INTO messages (conversation_id, topic_id, question_number, author, content, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6) 
			RETURNING id
			`

		err := repo.DB.QueryRow(query,
			conversation.ID,
			conversation.CurrentTopic,
			message.QuestionNumber,
			message.Author,
			message.Content,
			time.Now().UTC(),
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

func (repo *Repository) AddMessage(conversationID, topic_id, questionNumber int, message Message) (int, error) {
	query := `
			INSERT INTO messages (conversation_id, topic_id, question_number, author, content, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6) 
			RETURNING question_number
			`

	err := repo.DB.QueryRow(query,
		conversationID,
		topic_id,
		questionNumber,
		message.Author,
		message.Content,
		time.Now().UTC(),
	).Scan(&questionNumber)
	if err == sql.ErrNoRows {
		return 0, err
	} else if err != nil {
		log.Printf("Error querying conversation: %v\n", err)
		return 0, err
	}

	return questionNumber, nil
}

func (repo *Repository) GetMessages(conversationID, topic_id, questionNumber int) ([]Message, error) {
	messages := make([]Message, 0)

	query := `
			SELECT conversation_id, topic_id, question_number, author, content, created_at
			FROM messages
			WHERE conversation_id = $1 and topic_id = $2 and question_number = $3
			`

	rows, err := repo.DB.Query(
		query,
		conversationID,
		topic_id,
		questionNumber)
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
			&message.TopicID,
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
