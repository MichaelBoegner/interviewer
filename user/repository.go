package user

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
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

		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			switch pgErr.Constraint {
			case "users_email_key":
				return 0, fmt.Errorf("%w: %s", ErrDuplicateEmail, pgErr.Message)
			case "users_username_key":
				return 0, fmt.Errorf("%w: %s", ErrDuplicateUsername, pgErr.Message)
			default:
				return 0, fmt.Errorf("%w: %s", ErrDuplicateUser, pgErr.Message)
			}
		}

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

func (repo *Repository) GetUserByEmail(email string) (*User, error) {
	var user = &User{}
	err := repo.DB.QueryRow("SELECT id, username, email FROM users WHERE email= $1", email).Scan(&user.ID, &user.Username, &user.Email)

	if err == sql.ErrNoRows {
		log.Printf("Email invalid: %v", err)
		return nil, err
	} else if err != nil {
		log.Printf("Error querying database: %v\n", err)
		return nil, err
	}

	return user, nil
}

func (repo *Repository) GetUserByCustomerID(customerID string) (*User, error) {
	var user = &User{}
	err := repo.DB.QueryRow("SELECT id, username, email FROM users WHERE billing_customer_id = $1", customerID).Scan(&user.ID, &user.Username, &user.Email)
	if err == sql.ErrNoRows {
		log.Printf("CustomerID invalid: %v", err)
		return nil, err
	} else if err != nil {
		log.Printf("Error querying database: %v\n", err)
		return nil, err
	}

	return user, nil
}

func (repo *Repository) UpdatePasswordByEmail(email string, password []byte) error {
	query := `
			UPDATE users
			SET password = $1, updated_at = $2
			WHERE email = $3
			`

	result, err := repo.DB.Exec(query,
		password,
		time.Now().UTC(),
		email,
	)
	if err != nil {
		log.Printf("UpdatePasswordByEmail exec failed: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("RowsAffected failed: %v", err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("Email doesn't exist")
		return nil
	}

	return nil
}

func (repo *Repository) AddCredits(userID, credits int, creditType string) error {
	var column string
	switch creditType {
	case "individual":
		column = "individual_credits"
	case "subscription":
		column = "subscription_credits"
	default:
		return fmt.Errorf("invalid credit type: %s", creditType)
	}

	query := fmt.Sprintf(`
		UPDATE users
		SET %s = %s + $1, updated_at = $2
		WHERE id = $3
	`, column, column)

	_, err := repo.DB.Exec(query,
		credits,
		time.Now().UTC(),
		userID,
	)
	if err != nil {
		log.Printf("AddCredits failed: %v", err)
		return err
	}

	return nil
}
