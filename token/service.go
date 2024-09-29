package token

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func CreateRefreshToken(repo *Repository, userID int) (string, error) {
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
