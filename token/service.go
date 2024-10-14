package token

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateRefreshToken(repo TokenRepo, userID int) (string, error) {
	now := time.Now()

	refreshLength := 32
	refreshBytes := make([]byte, refreshLength)
	_, err := rand.Read([]byte(refreshBytes))
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(refreshBytes)
	expiry := time.Now().Add(time.Duration(168*60) * time.Hour)

	refreshToken := &RefreshToken{
		UserID:       userID,
		RefreshToken: token,
		ExpiresAt:    expiry,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = repo.AddRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	return refreshToken.RefreshToken, nil
}

func GetStoredRefreshToken(repo TokenRepo, userID int) (string, error) {
	storedToken, err := repo.GetStoredRefreshToken(userID)
	if err != nil {
		return "", err
	}
	return storedToken, nil
}

func VerifyRefreshToken(storedToken, providedToken string) bool {
	return subtle.ConstantTimeCompare([]byte(storedToken), []byte(providedToken)) == 1
}

func ExtractUserIDFromToken(tokenString string) (int, error) {
	jwtSecret := os.Getenv("JWT_SECRET")

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(tokenString *jwt.Token) (interface{}, error) {
		if _, ok := tokenString.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, fmt.Errorf("Unauthorized")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if ok && token.Valid {
		userID, err := strconv.Atoi(claims.UserID)
		if err != nil {
			return 0, err
		}

		return userID, nil
	}

	return 0, fmt.Errorf("Unauthorized")
}

func CreateJWT(id, expires int) (string, error) {
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
