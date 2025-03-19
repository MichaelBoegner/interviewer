package user

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(repo UserRepo, username, email, password string) (*User, error) {
	now := time.Now()

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

	err = repo.CreateUser(user)
	if err != nil {

		return nil, err
	}

	return user, nil
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

	jwToken, err := createJWT(id, 0)
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
		log.Printf("GetUser from database failed due to: %v", err)
		return nil, err
	}
	return userReturned, nil
}

func createJWT(id, expires int) (string, error) {
	var (
		key   []byte
		token *jwt.Token
	)

	jwtSecret := os.Getenv("JWT_SECRET")
	now := time.Now()
	if expires == 0 {
		expires = 3600
	}
	expiresAt := time.Now().Add(time.Duration(expires) * time.Second)
	key = []byte(jwtSecret)
	claims := jwt.RegisteredClaims{
		Issuer:    "interviewer",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Subject:   strconv.Itoa(id),
	}
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(key)
	if err != nil {
		log.Fatalf("Bad SignedString: %s", err)
		return "", err
	}

	return signedToken, nil
}
