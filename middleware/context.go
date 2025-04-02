package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/michaelboegner/interviewer/conversation"
)

type AcceptedVals struct {
	UserID         int                        `json:"user_id,omitempty"`
	InterviewID    int                        `json:"interview_id,omitempty"`
	ConversationID int                        `json:"conversation_id,omitempty"`
	Username       string                     `json:"username,omitempty"`
	Password       string                     `json:"password,omitempty"`
	Email          string                     `json:"email,omitempty"`
	AccessToken    string                     `json:"access_token,omitempty"`
	Message        string                     `json:"message,omitempty"`
	Conversation   *conversation.Conversation `json:"conversation,omitempty"`
}

type returnVals struct {
	Error string `json:"error,omitempty"`
}

type ContextKey string

const (
	ContextKeyParams      ContextKey = "params"
	ContextKeyTokenKey    ContextKey = "tokenKey"
	ContextKeyTokenParams ContextKey = "tokenParams"
)

func GetContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			userID   int
			err      error
			tokenKey string
		)
		tokenParts := strings.Split(r.Header.Get("Authorization"), " ")

		if len(tokenParts) < 2 {
			tokenKey = ""
		} else {
			tokenKey = tokenParts[1]
		}

		if tokenKey == "" {
			log.Printf("Token missing")
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if isAccessToken(tokenKey) {
			userID, err = VerifyToken(tokenKey)
			if err != nil {
				log.Printf("Supplied token returns error: %v", err)
				respondWithError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyTokenKey, tokenKey)
		ctx = context.WithValue(ctx, ContextKeyTokenParams, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isAccessToken(tokenString string) bool {
	return strings.Count(tokenString, ".") == 2
}

func VerifyToken(tokenString string) (int, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Printf("JWT secret is not set")
		err := errors.New("jwt secret is not set")
		return 0, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		userID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			return 0, err
		}
		return userID, nil
	} else {
		return 0, errors.New("Invalid token")
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	if code != 0 {
		w.WriteHeader(code)
	}

	respBody := returnVals{
		Error: msg,
	}
	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Write(data)
}
