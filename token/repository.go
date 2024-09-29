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
		"INSERT INTO users (user_id, refresh_token, expires_at, created_at, updated_at) "+
			"VALUES ($1, $2, $3, $4, $5)",
		token.Id,
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
