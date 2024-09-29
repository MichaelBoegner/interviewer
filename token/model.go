package token

import "time"

type RefreshToken struct {
	Id           int
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
