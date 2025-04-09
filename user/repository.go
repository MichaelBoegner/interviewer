package user

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

func (repo *Repository) CreateUser(user *User) (int, error) {
	var id int
	now := time.Now().UTC()

	query := `
		INSERT INTO users (username, password, email, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := repo.DB.QueryRow(query,
		user.Username,
		user.Password,
		user.Email,
		now,
		now,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, err
	} else if err != nil {
		log.Printf("CreateUser failed: %v\n", err)
		return 0, err
	}

	return id, nil
}

func (repo *Repository) GetPasswordandID(username string) (int, string, error) {
	var hashedPassword string
	var id int
	err := repo.DB.QueryRow("SELECT id, password from users WHERE username = $1",
		username,
	).Scan(&id, &hashedPassword)
	if err == sql.ErrNoRows {
		log.Printf("Username invalid: %v", err)
		return 0, "", err
	} else if err != nil {
		log.Printf("Error querying database: %v\n", err)
		return 0, "", err
	}

	return id, hashedPassword, nil
}

func (repo *Repository) GetUser(user *User) (*User, error) {
	err := repo.DB.QueryRow("SELECT id, username, email FROM users WHERE id= $1", user.ID).Scan(&user.ID, &user.Username, &user.Email)

	if err == sql.ErrNoRows {
		log.Printf("UserID invalid: %v", err)
		return nil, err
	} else if err != nil {
		log.Printf("Error querying database: %v\n", err)
		return nil, err
	}

	return user, nil
}
