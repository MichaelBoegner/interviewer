package token

import (
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	TokenRepo TokenRepo
	Logger    *slog.Logger
}

type RefreshToken struct {
	UserID       int
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewTokenService(tokenRepo TokenRepo, logger *slog.Logger) *TokenService {
	return &TokenService{
		TokenRepo: tokenRepo,
		Logger:    logger,
	}
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
