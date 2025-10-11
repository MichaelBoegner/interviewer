package user

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (u *UserService) VerificationToken(email, username, password string) (string, error) {
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

func (u *UserService) CreateUser(tokenStr string) (*User, error) {
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

	id, err := u.UserRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}
	user.ID = id
	return user, nil
}

func (u *UserService) LoginUser(email, password string) (string, int, error) {
	userID, hashedPassword, err := u.UserRepo.GetPasswordandID(email)
	if err != nil {
		return "", 0, err
	}

	user, err := u.UserRepo.GetUser(userID)
	if err != nil {
		log.Printf("u.UserRepo.GetUser failed: %v", err)
		return "", 0, err
	}

	if user.AccountStatus != "active" {
		return "", 0, ErrAccountDeleted
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return "", 0, err
	}

	return user.Username, userID, nil
}

func (u *UserService) GetUser(userID int) (*User, error) {
	userReturned, err := u.UserRepo.GetUser(userID)
	if err != nil {
		log.Printf("GetUser failed: %v", err)
		return nil, err
	}
	return userReturned, nil
}

func (u *UserService) MarkUserDeleted(userId int) error {
	err := u.UserRepo.MarkUserDeleted(userId)
	if err != nil {
		log.Printf("u.UserRepo.DeleteUser failed: %v", err)
		return err
	}

	return nil
}

func (u *UserService) GetUserByEmail(email string) error {
	_, err := u.UserRepo.GetUserByEmail(email)
	if err != nil {
		log.Printf("u.UserRepo.GetUserByEmail failed: %v", err)
		return err
	}

	return nil
}

func (u *UserService) ResetPassword(newPassword string, resetJWT string) error {
	email, err := verifyResetToken(resetJWT)
	if err != nil {
		return err
	}

	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.MinCost)
	if err != nil {
		log.Printf("GenerateFromPassword failed: %v\n", err)
		return err
	}

	err = u.UserRepo.UpdatePasswordByEmail(email, passwordHashed)
	if err != nil {
		log.Printf("UpdatePasswordByEmail failed: %v", err)
		return err
	}

	return nil
}

func (u *UserService) GetOrCreateByEmail(email, username string) (*User, error) {
	user, err := u.UserRepo.GetUserByEmail(email)
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

	id, err := u.UserRepo.CreateUser(newUser)
	if err != nil {
		log.Printf("CreateUser failed: %v", err)
		return nil, err
	}

	newUser.ID = id
	return newUser, nil
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
