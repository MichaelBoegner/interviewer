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
	err := repo.DB.QueryRow("SELECT refresh_token from refresh_tokens WHERE user_id = $1",
		userID,
	).Scan(&storedToken)
	if err == sql.ErrNoRows {
		log.Printf("User ID invalid: %v", err)
		return "", err
	} else if err != nil {
		log.Printf("Error querying database: %v\n", err)
		return "", err
	}
	return storedToken, nil
}