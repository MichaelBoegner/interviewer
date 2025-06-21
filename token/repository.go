package token

import (
	"database/sql"
	"log"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

// TODO: These are extremely simplistic in protocol (not tracking revoked, multi-sessions, etc. . . )
// Will advance this after unit tests, CI/CD, and AWS deploy.
func (repo *Repository) AddRefreshToken(token *RefreshToken) error {
	_, err := repo.DB.Exec(
		"INSERT INTO refresh_tokens (user_id, refresh_token, expires_at, created_at, updated_at) "+
			"VALUES ($1, $2, $3, $4, $5)",
		token.UserID,
		token.RefreshToken,
		token.ExpiresAt,
		token.CreatedAt,
		token.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}

	return nil
}

func (repo *Repository) GetStoredRefreshToken(userID int) (string, error) {
	var storedToken string
	err := repo.DB.QueryRow(`
		SELECT refresh_token FROM refresh_tokens 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT 1
		`, userID).Scan(&storedToken)
	if err == sql.ErrNoRows {
		log.Printf("User ID invalid: %v", err)
		return "", err
	} else if err != nil {
		log.Printf("Error querying database: %v\n", err)
		return "", err
	}
	return storedToken, nil
}

func (repo *Repository) DeleteRefreshToken(userID int) error {
	_, err := repo.DB.Exec(`
		DELETE FROM refresh_tokens
		WHERE user_id = $1
	`, userID)

	if err != nil {
		log.Printf("Failed to delete refresh tokens for user %d: %v", userID, err)
		return err
	}

	return nil
}
