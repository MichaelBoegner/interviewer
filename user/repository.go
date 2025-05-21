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

func (repo *Repository) GetPasswordandID(email string) (int, string, error) {
	var hashedPassword string
	var id int
	err := repo.DB.QueryRow("SELECT id, password from users WHERE email = $1",
		email,
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

func (repo *Repository) GetUser(userID int) (*User, error) {
	user := &User{}

	err := repo.DB.QueryRow(`SELECT id, username, email, individual_credits, subscription_credits, subscription_end_date, subscription_status
							FROM users 
							WHERE id= $1`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.IndividualCredits,
		&user.SubscriptionCredits,
		&user.SubscriptionEndDate,
		&user.SubscriptionStatus,
	)

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
	err := repo.DB.QueryRow("SELECT id, username, email, subscription_tier FROM users WHERE email= $1", email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.SubscriptionTier,
	)

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
		WHERE id = $3 AND %s + $1 >= 0
	`, column, column, column)

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

func (repo *Repository) UpdateSubscriptionData(userID int, status, tier string, startsAt, endsAt time.Time) error {
	query := `
		UPDATE users
		SET subscription_status = $1,
		    subscription_tier = $2,
		    subscription_start_date = $3,
		    subscription_end_date = $4,
		    updated_at = $5
		WHERE id = $6
	`

	_, err := repo.DB.Exec(query,
		status,
		tier,
		startsAt.UTC(),
		endsAt.UTC(),
		time.Now().UTC(),
		userID,
	)
	if err != nil {
		log.Printf("UpdateSubscriptionData failed: %v", err)
		return err
	}

	return nil
}

func (repo *Repository) UpdateSubscriptionStatusData(userID int, status string) error {
	query := `
		UPDATE users
		SET subscription_status = $1,
		    updated_at = $2
		WHERE id = $3
	`

	_, err := repo.DB.Exec(query,
		status,
		time.Now().UTC(),
		userID,
	)
	if err != nil {
		log.Printf("CancelSubscriptionData failed: %v", err)
		return err
	}

	return nil
}
