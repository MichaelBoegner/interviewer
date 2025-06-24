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

func VerificationToken(email, username, password string) (string, error) {
	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"email":         email,
		"username":      username,
		"password_hash": string(passwordHashed),
		"purpose":       "verify_email",
		"exp":           time.Now().Add(15 * time.Minute).Unix(),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func CreateUser(repo UserRepo, tokenStr string) (*User, error) {
	claims := &EmailClaims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !tkn.Valid {
		return nil, errors.New("invalid or expired token")
	}

	if claims.Purpose != "verify_email" {
		return nil, errors.New("token used for wrong purpose")
	}

	user := &User{
		Email:             claims.Email,
		Username:          claims.Username,
		Password:          []byte(claims.PasswordHash),
		IndividualCredits: 1,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	id, err := repo.CreateUser(user)
	if err != nil {
		return nil, err
	}
	user.ID = id
	return user, nil
}

func LoginUser(repo UserRepo, email, password string) (string, string, int, error) {
	userID, hashedPassword, err := repo.GetPasswordandID(email)
	if err != nil {
		return "", "", 0, err
	}

	user, err := repo.GetUser(userID)
	if err != nil {
		log.Printf("repo.GetUser failed: %v", err)
		return "", "", 0, err
	}

	if user.AccountStatus != "active" {
		return "", "", 0, ErrAccountDeleted
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return "", "", 0, err
	}

	jwToken, err := token.CreateJWT(strconv.Itoa(userID), 0)
	if err != nil {
		log.Printf("JWT creation failed: %v", err)
		return "", "", 0, err
	}

	return jwToken, user.Username, userID, nil
}

func GetUser(repo UserRepo, userID int) (*User, error) {
	userReturned, err := repo.GetUser(userID)
	if err != nil {
		log.Printf("GetUser failed: %v", err)
		return nil, err
	}
	return userReturned, nil
}

func MarkUserDeleted(repo UserRepo, userId int) error {
	err := repo.MarkUserDeleted(userId)
	if err != nil {
		log.Printf("repo.DeleteUser failed: %v", err)
		return err
	}

	return nil
}

func GetUserByEmail(repo UserRepo, email string) error {
	_, err := repo.GetUserByEmail(email)
	if err != nil {
		log.Printf("repo.GetUserByEmail failed: %v", err)
		return err
	}

	return nil
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

func GetOrCreateByEmail(repo UserRepo, email, username string) (*User, error) {
	user, err := repo.GetUserByEmail(email)
	if err == nil {
		return user, nil
	}

	newUser := &User{
		Email:             email,
		Username:          username,
		Password:          []byte("github_login"),
		AccountStatus:     "active",
		IndividualCredits: 1,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	id, err := repo.CreateUser(newUser)
	if err != nil {
		log.Printf("CreateUser failed: %v", err)
		return nil, err
	}

	newUser.ID = id
	return newUser, nil
}
