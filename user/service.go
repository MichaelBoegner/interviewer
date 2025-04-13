package user

import (
	"log"
	"time"

	"github.com/michaelboegner/interviewer/token"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(repo UserRepo, username, email, password string) (*User, error) {
	now := time.Now().UTC()

	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, err
	}

	user := &User{
		Username:  username,
		Email:     email,
		Password:  []byte(passwordHashed),
		CreatedAt: now,
		UpdatedAt: now,
	}

	id, err := repo.CreateUser(user)
	if err != nil {
		log.Printf("CreateUser failed: %v", err)
		return nil, err
	}
	user.ID = id

	return user, nil
	// For preventing user creation in deployment:
	// err := errors.New("User creation has been disabled for now. Live demos available upon request!")
	// return nil, err
}

func LoginUser(repo UserRepo, username, password string) (string, int, error) {
	id, hashedPassword, err := repo.GetPasswordandID(username)
	if err != nil {
		return "", 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return "", 0, err
	}

	jwToken, err := token.CreateJWT(id, 0)
	if err != nil {
		log.Printf("JWT creation failed: %v", err)
		return "", 0, err
	}

	return jwToken, id, nil
}

func GetUser(repo UserRepo, userID int) (*User, error) {
	user := &User{
		ID: userID,
	}

	userReturned, err := repo.GetUser(user)
	if err != nil {
		log.Printf("GetUser failed: %v", err)
		return nil, err
	}
	return userReturned, nil
}
