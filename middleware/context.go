package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type AcceptedVals struct {
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	AccessToken string `json:"access_token"`
}

type returnVals struct {
	Error string `json:"error,omitempty"`
}

func GetContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		tokenParts := strings.Split(r.Header.Get("Authorization"), " ")
		var tokenPart string
		if len(tokenParts) < 2 {
			tokenPart = ""
		} else {
			tokenPart = tokenParts[1]
		}

		// Read the request body into a byte slice
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Recreate the request body for the handler (as it is read-only once consumed)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		// Decode the body into AcceptedVals struct
		var params AcceptedVals
		if err := json.Unmarshal(body, &params); err != nil {
			respondWithError(w, http.StatusBadRequest, "Error decoding params")
			return
		}

		// Set extracted data in context for access by handlers
		ctx := r.Context()
		ctx = context.WithValue(ctx, "token", tokenPart)
		ctx = context.WithValue(ctx, "params", params)

		// Pass along the request with the new context to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
