package user

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
		Password:  passwordHashed,
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
	// err := errors.New("We are not quite yet fully live. Check back again in the future!")
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

	jwToken, err := token.CreateJWT(strconv.Itoa(id), 0)
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

func RequestPasswordReset(repo UserRepo, email string) (string, error) {
	user, err := repo.GetUserByEmail(email)
	if err != nil {
		log.Printf("GetUserByEmail failed: %v", err)
		return "", err
	}

	resetJWT, err := token.CreateJWT(user.Email, 900)
	if err != nil {
		log.Printf("CreateJWT failed: %v", err)
		return "", err
	}

	return resetJWT, nil
}

func ResetPassword(repo UserRepo, newPassword string, resetJWT string) error {
	email, err := verifyResetToken(resetJWT)
	if err != nil {
		return err
	}

	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.MinCost)
	if err != nil {
		log.Printf("GenerateFromPassword failed: %v\n", err)
		return err
	}

	err = repo.UpdatePasswordByEmail(email, passwordHashed)
	if err != nil {
		log.Printf("UpdatePasswordByEmail failed: %v", err)
		return err
	}

	return nil
}

func verifyResetToken(tokenString string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Printf("JWT secret is not set")
		err := errors.New("jwt secret is not set")
		return "", err
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims.Subject, nil
	} else {
		return "", errors.New("Invalid token")
	}
}
