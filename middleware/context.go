package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/michaelboegner/interviewer/conversation"
)

type AcceptedVals struct {
	UserID       int                        `json:"user_id"`
	Username     string                     `json:"username"`
	Password     string                     `json:"password"`
	Email        string                     `json:"email"`
	AccessToken  string                     `json:"access_token"`
	InterviewID  int                        `json:"interview_id"`
	Conversation *conversation.Conversation `json:"conversation,omitempty"`
	Message      *conversation.Message      `json:"message,omitempty"`
}

type returnVals struct {
	Error string `json:"error,omitempty"`
}

type UpdateConversation struct {
	ConversationID int                   `json:"conversation_id"`
	TopicID        int                   `json:"topic_id"`
	QuestionID     int                   `json:"question_id"`
	QuestionNumber int                   `json:"question_number"`
	Message        *conversation.Message `json:"message"`
}

type ContextKey string

const (
	ContextKeyParams   ContextKey = "params"
	ContextKeyTokenKey ContextKey = "tokenKey"
)

func GetContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		tokenParts := strings.Split(r.Header.Get("Authorization"), " ")

		var tokenKey string
		if len(tokenParts) < 2 {
			tokenKey = ""
		} else {
			tokenKey = tokenParts[1]
		}

		if isAccessToken(tokenKey) {
			_, err := VerifyToken(tokenKey)
			if err != nil {
				log.Printf("Supplied token returns error: %v", err)
				respondWithError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
		}

		// Read the request body into a byte slice
		body, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Recreate the request body for the handler (as it is read-only once consumed)
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Decode the body into UpdateConversation or AcceptedVals struct
		var params interface{}
		if strings.Contains(r.URL.Path, "/conversations") {
			var updateParams UpdateConversation
			err := json.Unmarshal(body, &updateParams)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "Error decoding update conversation params")
				return
			}

			params = updateParams
		} else {
			var generalParams AcceptedVals
			err := json.Unmarshal(body, &generalParams)
			if err != nil {
				log.Printf("Unmarshal error: %v", err)
				log.Printf("Raw body: %s", string(body))
				respondWithError(w, http.StatusBadRequest, "Error decoding general params")
				return
			}

			params = generalParams
		}

		// Set extracted data in context for access by handlers
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyTokenKey, tokenKey)
		ctx = context.WithValue(ctx, ContextKeyParams, params)

		// Pass along the request with the new context to the handler
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
