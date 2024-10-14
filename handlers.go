package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

type returnVals struct {
	Error        string            `json:"error,omitempty"`
	ID           int               `json:"id,omitempty"`
	UserID       int               `json:"user_id,omitempty"`
	Body         string            `json:"body,omitempty"`
	Username     string            `json:"username,omitempty"`
	Email        string            `json:"email,omitempty"`
	Token        string            `json:"token,omitempty"`
	Users        map[int]user.User `json:"users,omitempty"`
	Questions    map[int]string    `json:"firstQuestion,omitempty"`
	JWToken      string            `json:"jwtoken,omitempty"`
	RefreshToken string            `json:"refresh_token,omitempty"`
}

func (apiCfg *apiConfig) usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST create a user
	case http.MethodPost:
		params := r.Context().Value("params").(middleware.AcceptedVals)

		user, err := user.CreateUser(apiCfg.UserRepo, params.Username, params.Email, params.Password)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		payload := &returnVals{
			Username: user.Username,
			Email:    user.Email,
		}
		respondWithJSON(w, 200, payload)

	case http.MethodGet:
		users, err := user.GetUsers(apiCfg.UserRepo)
		if err != nil {
			log.Printf("GetUsers failed due to: %v", err)
		}

		payload := &returnVals{
			Users: users.Users,
		}

		respondWithJSON(w, http.StatusOK, payload)
	}
}

func (apiCfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST login a user
	case http.MethodPost:
		params := r.Context().Value("params").(middleware.AcceptedVals)

		jwToken, userID, err := user.LoginUser(apiCfg.UserRepo, params.Username, params.Password)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid username or password.")
		}

		refreshToken, err := token.CreateRefreshToken(apiCfg.TokenRepo, userID)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "")
		}

		payload := returnVals{
			UserID:       userID,
			Username:     params.Username,
			JWToken:      jwToken,
			RefreshToken: refreshToken,
		}

		respondWithJSON(w, http.StatusOK, payload)
	}
}

func (apiCfg *apiConfig) interviewsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST start a resource instance of an interview and return the first question
	case http.MethodPost:
		interviewStarted, err := interview.StartInterview(apiCfg.InterviewRepo, 1, 30, 3, "easy")
		if err != nil {
			log.Printf("Interview failed to start: %v", err)
		}

		payload := returnVals{
			Questions: interviewStarted.Questions,
		}

		respondWithJSON(w, http.StatusOK, payload)
	}
}

func (apiCfg *apiConfig) refreshTokensHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST generate and return userID and a refreshToken
	case http.MethodPost:
		providedToken := r.Context().Value("token").(string)
		params := r.Context().Value("params").(middleware.AcceptedVals)

		storedToken, err := token.GetStoredRefreshToken(apiCfg.TokenRepo, params.UserID)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "User ID is invalid.")
		}

		ok := token.VerifyRefreshToken(storedToken, providedToken)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "Refresh token is invalid.")
		}

		refreshToken, err := token.CreateRefreshToken(apiCfg.TokenRepo, params.UserID)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "")
		}

		jwToken, err := token.CreateJWT(params.UserID, 0)
		if err != nil {
			log.Printf("JWT creation failed: %v", err)
			respondWithError(w, http.StatusInternalServerError, "")
		}

		payload := &returnVals{
			ID:           params.UserID,
			JWToken:      jwToken,
			RefreshToken: refreshToken,
		}
		respondWithJSON(w, http.StatusAccepted, payload)
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	if code != 0 {
		w.WriteHeader(code)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
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
