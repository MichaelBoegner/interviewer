package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RefreshToken struct {
	UserID       int
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CustomClaims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

func (c *CustomClaims) GetAudience() (jwt.ClaimStrings, error) {
	return c.Audience, nil
}

type TokenRepo interface {
	AddRefreshToken(token *RefreshToken) error
	GetStoredRefreshToken(userID int) (string, error)
	DeleteRefreshToken(userID int) error
}
