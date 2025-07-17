package embedding

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type PGRepository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *PGRepository {
	return &PGRepository{DB: db}
}

func (r *PGRepository) StoreEmbedding(ctx context.Context, e Embedding) error {
	query := `
		INSERT INTO conversation_embeddings (
			interview_id,
			conversation_id,
			message_id,
			topic_id,
			question_number,
			summary,
			embedding,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (interview_id, conversation_id, topic_id, question_number, message_id) DO NOTHING;
	`

	_, err := r.DB.ExecContext(ctx, query,
		e.InterviewID,
		e.ConversationID,
		e.MessageID,
		e.TopicID,
		e.QuestionNumber,
		e.Summary,
		pq.Array(e.Vector),
		e.CreatedAt,
	)
	return err
}

func (r *PGRepository) GetSimilarEmbeddings(
	ctx context.Context,
	interviewID, topicID, questionNumber, excludeMessageID int,
	queryVec []float32,
	limit int,
) ([]string, error) {
	query := `
		SELECT summary
		FROM conversation_embeddings
		WHERE interview_id = $1
		  AND topic_id = $2
		  AND question_number = $3
		  AND message_id != $4
		ORDER BY embedding <-> $5
		LIMIT $6;
	`

	rows, err := r.DB.QueryContext(ctx, query,
		interviewID,
		topicID,
		questionNumber,
		excludeMessageID,
		pq.Array(queryVec),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var summaries []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		summaries = append(summaries, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return summaries, nil
}
