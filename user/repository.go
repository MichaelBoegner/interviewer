package user

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

func (repo *Repository) CreateUser(user *User) error {
	_, err := repo.DB.Exec(
		"INSERT INTO users (username, password, email, created_at, updated_at) "+
			"VALUES ($1, $2, $3, $4, $5)",
		user.Username,
		user.Password,
		user.Email,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}

	return nil
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

func (repo *Repository) GetUsers(users *Users) (*Users, error) {
	rows, err := repo.DB.Query("SELECT id, username, email FROM users")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		user := User{}
		rows.Scan(&user.ID, &user.Username, &user.Email)
		users.Users[user.ID] = user
	}
	return users, nil
}
