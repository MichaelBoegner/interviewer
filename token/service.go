package token

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (t *TokenService) CreateJWT(subject string, expires int) (string, error) {
	var (
		key     []byte
		jwtoken *jwt.Token
	)

	jwtSecret := os.Getenv("JWT_SECRET")
	now := time.Now().UTC()
	if expires == 0 {
		expires = 36000
	}
	expiresAt := now.Add(time.Duration(expires) * time.Second)
	key = []byte(jwtSecret)
	claims := jwt.RegisteredClaims{
		Issuer:    "interviewer",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Subject:   subject,
	}
	jwtoken = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtoken.SignedString(key)
	if err != nil {
		t.Logger.Error("Bad SignedString", "error", err)
		return "", err
	}

	return tokenString, nil
}

func (t *TokenService) CreateRefreshToken(userID int) (string, error) {
	now := time.Now().UTC()

	refreshLength := 32
	refreshBytes := make([]byte, refreshLength)
	_, err := rand.Read([]byte(refreshBytes))
	if err != nil {
		t.Logger.Error("rand.Read failed", "error", err)
		return "", err
	}
	token := hex.EncodeToString(refreshBytes)
	expiry := now.Add(time.Duration(168*60) * time.Hour)

	refreshToken := &RefreshToken{
		UserID:       userID,
		RefreshToken: token,
		ExpiresAt:    expiry,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = t.TokenRepo.AddRefreshToken(refreshToken)
	if err != nil {
		t.Logger.Error("t.TokenRepo.AddRefreshToken failed", "error", err)
		return "", err
	}

	return refreshToken.RefreshToken, nil
}

func (t *TokenService) DeleteRefreshToken(userID int) error {
	err := t.TokenRepo.DeleteRefreshToken(userID)
	if err != nil {
		t.Logger.Error("t.TokenRepo.DeleteRefreshToken failed", "error", err)
		return err
	}

	return nil
}

func (t *TokenService) GetStoredRefreshToken(userID int) (string, error) {
	storedToken, err := t.TokenRepo.GetStoredRefreshToken(userID)
	if err != nil {
		t.Logger.Error("t.TokenRepo.GetStoredRefreshToken failed", "error", err)
		return "", err
	}
	return storedToken, nil
}

func (t *TokenService) VerifyRefreshToken(storedToken, providedToken string) bool {
	return subtle.ConstantTimeCompare([]byte(storedToken), []byte(providedToken)) == 1
}

func (t *TokenService) ExtractUserIDFromToken(tokenString string) (int, error) {
	jwtSecret := os.Getenv("JWT_SECRET")

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(tokenString *jwt.Token) (interface{}, error) {
		if _, ok := tokenString.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, fmt.Errorf("unauthorized")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		t.Logger.Error("jwt.ParseWithClaims failed", "error", err)
		return 0, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if ok && token.Valid {
		userID, err := strconv.Atoi(claims.UserID)
		if err != nil {
			t.Logger.Error("strconv.Atoi failed", "error", err)
			return 0, err
		}

		return userID, nil
	}

	return 0, fmt.Errorf("unauthorized")
}
