package user

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

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
