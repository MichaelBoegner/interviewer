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
	now := time.Now().UTC()

	var dummy int
	checkQuery := `
		SELECT 1 FROM users
		WHERE account_status = 'deleted'
		AND email LIKE '%' || $1
		LIMIT 1
	`
	err := repo.DB.QueryRow(checkQuery, user.Email).Scan(&dummy)
	if err == sql.ErrNoRows {
		user.IndividualCredits = 1
	} else if err != nil {
		log.Printf("Email reuse check failed: %v\n", err)
		return 0, err
	} else {
		user.IndividualCredits = 0
	}

	var id int
	insertQuery := `
		INSERT INTO users (username, password, email, individual_credits, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err = repo.DB.QueryRow(insertQuery,
		user.Username,
		user.Password,
		user.Email,
		user.IndividualCredits,
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
			default:
				return 0, fmt.Errorf("%w: %s", ErrDuplicateUser, pgErr.Message)
			}
		}

		log.Printf("CreateUser failed: %v\n", err)
		return 0, err
	}

	return id, nil
}

func (repo *Repository) MarkUserDeleted(userID int) error {
	query := `
		UPDATE users
		SET 
			account_status = 'deleted',
			email = CONCAT('deleted_', id, '_', email),
			username = CONCAT('deleted_user_', id),
			password = '',
			subscription_id = '',
			updated_at = $1
		WHERE id = $2
	`

	_, err := repo.DB.Exec(query, time.Now().UTC(), userID)
	if err != nil {
		log.Printf("MarkUserDeleted failed: %v\n", err)
		return err
	}

	return nil
}

func (repo *Repository) GetPasswordandID(email string) (int, string, error) {
	var hashedPassword string
	var id int
	err := repo.DB.QueryRow("SELECT id, password from users WHERE email = $1",
		email,
	).Scan(&id, &hashedPassword)
	if err == sql.ErrNoRows {
		log.Printf("email invalid: %v", err)
		return 0, "", err
	} else if err != nil {
		log.Printf("Error querying database: %v\n", err)
		return 0, "", err
	}

	return id, hashedPassword, nil
}

func (repo *Repository) GetUser(userID int) (*User, error) {
	user := &User{}

	err := repo.DB.QueryRow(`SELECT 
								id, 
								username, 
								email, 
								individual_credits, 
								subscription_credits, 
								subscription_start_date, 
								subscription_end_date, 
								subscription_status, 
								subscription_tier, 
								subscription_id,
								account_status
							FROM users 
							WHERE id= $1`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.IndividualCredits,
		&user.SubscriptionCredits,
		&user.SubscriptionStartDate,
		&user.SubscriptionEndDate,
		&user.SubscriptionStatus,
		&user.SubscriptionTier,
		&user.SubscriptionID,
		&user.AccountStatus,
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
	err := repo.DB.QueryRow(`SELECT 
								id, 
								username, 
								email, 
								subscription_tier, 
								subscription_credits
							FROM users 
							WHERE email= $1`, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.SubscriptionTier,
		&user.SubscriptionCredits,
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

	if credits < 0 {
		var current int
		err := repo.DB.QueryRow(fmt.Sprintf("SELECT %s FROM users WHERE id = $1", column), userID).Scan(&current)
		if err != nil {
			log.Printf("repo.DB.QueryRow failed: %v", err)
			return fmt.Errorf("fetch current credit balance failed: %w", err)
		}

		if current < -credits {
			credits = -current
		}
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

func (repo *Repository) UpdateSubscriptionData(userID int, status, tier, subscriptionID string, startsAt, endsAt time.Time) error {
	query := `
		UPDATE users
		SET subscription_status = $1,
		    subscription_tier = $2,
			subscription_id = $3,
		    subscription_start_date = $4,
		    subscription_end_date = $5,
		    updated_at = $6
		WHERE id = $7
	`

	_, err := repo.DB.Exec(query,
		status,
		tier,
		subscriptionID,
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

func (repo *Repository) HasActiveOrCancelledSubscription(email string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM users
			WHERE email = $1 AND subscription_tier != 'free' AND subscription_status IN ('active', 'cancelled')
		)
	`
	var exists bool
	err := repo.DB.QueryRow(query, email).Scan(&exists)
	if err != nil {
		log.Printf("repo.DB.QueryRow failed: %v", err)
		return exists, err
	}

	return exists, err
}
