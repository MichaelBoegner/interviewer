package billing

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

func (r *Repository) LogCreditTransaction(tx CreditTransaction) error {
	query := `
		INSERT INTO credit_transactions (user_id, amount, credit_type, reason, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.DB.Exec(query,
		tx.UserID,
		tx.Amount,
		tx.CreditType,
		tx.Reason,
		time.Now().UTC(),
	)
	if err != nil {
		log.Printf("LogCreditTransaction failed: %v", err)
		return err
	}

	return nil
}

func (r *Repository) HasWebhookBeenProcessed(webhookID string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM processed_webhooks WHERE webhook_id = $1)
    `, webhookID).Scan(&exists)
	return exists, err
}

func (r *Repository) MarkWebhookProcessed(webhookID string, event string) error {
	_, err := r.DB.Exec(`
        INSERT INTO processed_webhooks (webhook_id, event_name, processed_at)
        VALUES ($1, $2, $3)
    `, webhookID, event, time.Now().UTC())

	return err
}
