package user

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(repo *Repository, username, email, password string) (*User, error) {
	now := time.Now()

	user := &User{
		Username:  username,
		Email:     email,
		Password:  []byte(password),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func LoginUser(repo *Repository, username, password string) (string, error) {
	id, hashedPassword, err := repo.GetPasswordandID(username)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return "", err
	}

	jwToken, err := createJWT(id, 0)
	if err != nil {
		log.Printf("JWT creation failed: %v", err)
		return "", err
	}

	return jwToken, nil
}

func GetUsers(repo *Repository) (*Users, error) {
	userMap := make(map[int]User)
	users := &Users{
		Users: userMap,
	}
	usersReturned, err := repo.GetUsers(users)
	if err != nil {
		log.Printf("GetUsers from database failed due to: %v", err)
		return nil, err
	}
	return usersReturned, nil
}

func createJWT(id, expires int) (string, error) {
	var (
		key []byte
		t   *jwt.Token
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
	t = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(key)
	if err != nil {
		log.Fatalf("Bad SignedString: %s", err)
		return "", err
	}

	return s, nil
}

func createRefreshToken() (string, time.Time, error) {
	refreshLength := 32
	refreshBytes := make([]byte, refreshLength)
	_, err := rand.Read([]byte(refreshBytes))
	if err != nil {
		return "", time.Time{}, err
	}
	refreshToken := hex.EncodeToString(refreshBytes)
	expiry := time.Now().Add(time.Duration(168*60) * time.Hour)

	return refreshToken, expiry, nil
}
